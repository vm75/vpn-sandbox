package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

type IpInfo map[string]interface{}

var (
	staticDir     = "./static"
	ipInfo        = &ipInfoCache{}
	statusUpdates = newStatusNotifier()
)

const statusHeartbeatInterval = 15 * time.Second

type ipInfoStatus struct {
	Output     IpInfo     `json:"output"`
	ExecutedAt *time.Time `json:"executedAt,omitempty"`
	Event      string     `json:"event"`
	EventAt    time.Time  `json:"eventAt"`
	Stale      bool       `json:"stale"`
}

type ipInfoCache struct {
	mu          sync.RWMutex
	refreshMu   sync.Mutex
	lookup      func(IpInfo) error
	output      IpInfo
	executedAt  time.Time
	lastEvent   string
	lastEventAt time.Time
}

type statusEventListener struct{}

type statusNotifier struct {
	mu          sync.RWMutex
	subscribers map[chan string]struct{}
}

func newStatusNotifier() *statusNotifier {
	return &statusNotifier{subscribers: make(map[chan string]struct{})}
}

func (n *statusNotifier) subscribe() (<-chan string, func()) {
	updates := make(chan string, 1)
	n.mu.Lock()
	n.subscribers[updates] = struct{}{}
	n.mu.Unlock()

	return updates, func() {
		n.mu.Lock()
		delete(n.subscribers, updates)
		n.mu.Unlock()
	}
}

func (n *statusNotifier) notify(event string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for subscriber := range n.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (i *ipInfoCache) markEvent(name string, at time.Time) {
	i.mu.Lock()
	i.lastEvent = name
	i.lastEventAt = at
	i.mu.Unlock()
}

func (i *ipInfoCache) store(output IpInfo, executedAt time.Time) {
	i.mu.Lock()
	i.output = output
	i.executedAt = executedAt
	i.mu.Unlock()
}

func (i *ipInfoCache) snapshot() ipInfoStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()

	output := make(IpInfo, len(i.output))
	for key, value := range i.output {
		output[key] = value
	}

	var executedAt *time.Time
	if !i.executedAt.IsZero() {
		value := i.executedAt
		executedAt = &value
	}

	return ipInfoStatus{
		Output:     output,
		ExecutedAt: executedAt,
		Event:      i.lastEvent,
		EventAt:    i.lastEventAt,
		Stale:      i.executedAt.IsZero() || i.executedAt.Before(i.lastEventAt),
	}
}

func (i *ipInfoCache) refresh() bool {
	i.refreshMu.Lock()
	defer i.refreshMu.Unlock()

	lookup := i.lookup
	if lookup == nil {
		lookup = func(output IpInfo) error {
			return utils.GetIpInfo(output)
		}
	}

	updated := IpInfo{}
	if err := lookup(updated); err != nil {
		return false
	}

	i.store(updated, time.Now())
	return true
}

func notifyStatus(event string) {
	statusUpdates.notify(event)
}

func (statusEventListener) HandleEvent(event utils.Event) {
	notifyStatus(event.Name)
}

type ModuleStatus struct {
	Running bool                   `json:"running"`
	Info    map[string]interface{} `json:"info"`
}

func queryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) == 0 {
			continue
		}
		params[k] = v[0]
	}
	return params
}

func getStatus() map[string]interface{} {
	globalConfig, _ := core.GetGlobalConfig()
	status := make(map[string]interface{})
	status["global"] = map[string]interface{}{
		"config": globalConfig,
	}

	for name, module := range core.GetModules() {
		moduleStatus := make(map[string]interface{})
		moduleStatus["running"] = module.IsRunning()
		moduleStatus["config"], _ = module.GetConfig(nil)
		status[name] = moduleStatus
	}

	status["ipInfo"] = ipInfo.snapshot()

	return status
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	updates, unsubscribe := statusUpdates.subscribe()
	defer unsubscribe()

	data, _ := json.Marshal(getStatus())
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(statusHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case event := <-updates:
			utils.LogLn("Received event:", event)
			data, _ := json.Marshal(getStatus())
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func forceRefreshHandler(w http.ResponseWriter, _ *http.Request) {
	ipInfo.markEvent("force", time.Now())
	if !ipInfo.refresh() {
		notifyStatus("force")
		http.Error(w, "failed to refresh IP info", http.StatusBadGateway)
		return
	}
	notifyStatus("ip-info")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ipInfo.snapshot())
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": core.Version})
}

func getGlobalConfigHandler(w http.ResponseWriter, r *http.Request) {
	config, err := core.GetGlobalConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(config)
}

func saveGlobalConfigHandler(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = core.SaveGlobalConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Get Module status
func getModuleStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	isRunning, err := core.GetModuleStatus(module)
	status := ModuleStatus{
		Running: isRunning,
		Info:    ipInfo.snapshot().Output,
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(status)
}

// Enable Module
func enableModuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	startNow := r.URL.Query().Get("start") == "true"

	err := core.EnableModule(module, startNow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// Disable Module
func disableModuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	stopNow := r.URL.Query().Get("stop") == "true"

	err := core.DisableModule(module, stopNow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// Restart Module
func restartModuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	err := core.RestartModule(module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func getModuleConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	params := queryParams(r)
	config, err := core.GetModuleConfig(module, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(config)
}

// Save a config (new or existing)
func saveModuleConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	var params = queryParams(r)
	var config map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = core.SaveModuleConfig(module, params, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type FileInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

// Helper function to sanitize paths
func sanitizePath(inputPath string) string {
	// Prevent navigating outside baseDir
	cleanPath := filepath.Clean("/" + inputPath) // Ensure the path starts with a slash
	return filepath.Join(core.VarDir, cleanPath)
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	absPath := sanitizePath(relPath)

	files, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".auth") {
			continue
		}
		fileInfos = append(fileInfos, FileInfo{
			Name:  file.Name(),
			Path:  filepath.Join(relPath, file.Name()), // Preserve the relative path for the client
			IsDir: file.IsDir(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfos)
}

func fileContentHandler(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	absPath := sanitizePath(relPath)

	// Ensure the requested path is inside the base directory
	if !strings.HasPrefix(absPath, core.VarDir) || strings.HasSuffix(absPath, ".auth") {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

// Separate function to handle static files
func handleStaticFiles(r *mux.Router) {
	// Serve static files from /static and root (/)
	fs := http.FileServer(http.Dir(staticDir))
	noCache := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache")
			handler.ServeHTTP(w, r)
		})
	}
	r.PathPrefix("/static/").Handler(noCache(http.StripPrefix("/static/", fs)))
	r.PathPrefix("/").Handler(noCache(http.StripPrefix("/", fs))) // Serve "/" from staticDir
}

func (i *ipInfoCache) HandleEvent(event utils.Event) {
	i.markEvent(event.Name, time.Now())
	notifyStatus(event.Name)

	if i.refresh() {
		notifyStatus("ip-info")
	}
}

func WebServer(port string) {
	utils.RegisterListener([]string{"vpn-up", "vpn-down"}, ipInfo)
	utils.RegisterListener([]string{"proxy-up", "proxy-down"}, statusEventListener{})

	ipInfo.markEvent("startup", time.Now())
	go func() {
		if ipInfo.refresh() {
			notifyStatus("ip-info")
		}
	}()

	// Create a new Gorilla Mux router
	r := mux.NewRouter()

	// Config-related routes
	r.HandleFunc("/api/status", statusHandler).Methods("GET")
	r.HandleFunc("/api/version", versionHandler).Methods("GET")
	r.HandleFunc("/api/force-refresh", forceRefreshHandler).Methods("GET")
	r.HandleFunc("/api/config", getGlobalConfigHandler).Methods("GET")
	r.HandleFunc("/api/config/save", saveGlobalConfigHandler).Methods("POST")

	// Handle file
	r.HandleFunc("/api/files", listFilesHandler)
	r.HandleFunc("/api/file", fileContentHandler)

	// Module
	r.HandleFunc("/api/{module}/status", getModuleStatusHandler).Methods("GET")
	r.HandleFunc("/api/{module}/enable", enableModuleHandler).Methods("POST")
	r.HandleFunc("/api/{module}/disable", disableModuleHandler).Methods("POST")
	r.HandleFunc("/api/{module}/restart", restartModuleHandler).Methods("POST")
	r.HandleFunc("/api/{module}/config", getModuleConfigHandler).Methods("GET")
	r.HandleFunc("/api/{module}/config/save", saveModuleConfigHandler).Methods("POST")

	// Custom module routes
	for _, module := range core.GetModules() {
		module.RegisterRoutes(r)
	}

	// Serve static files
	handleStaticFiles(r)

	// Start the server
	utils.LogF("Server starting on port %s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		utils.LogFatal(err)
	}
}
