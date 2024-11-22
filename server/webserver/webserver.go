package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

type IpInfo map[string]interface{}

var (
	staticDir        = "./static"
	ipInfo           = IpInfo{}
	nwChangedChannel = make(chan string) // Channel for sending status updates
)

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

	utils.GetIpInfo(ipInfo)
	status["ipInfo"] = ipInfo

	return status
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel to close the connection on client disconnect
	clientGone := r.Context().Done()

	data, _ := json.Marshal(getStatus())
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush() // Ensure data is sent immediately

	for {
		select {
		case event := <-nwChangedChannel: // Receive new status from the channel
			utils.LogLn("Received event:", event)
			data, _ := json.Marshal(getStatus())
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush() // Ensure data is sent immediately
		case <-clientGone: // Client disconnected
			return
		}
	}
}

func forceRefreshHandler(w http.ResponseWriter, _ *http.Request) {
	nwChangedChannel <- "force"
	w.WriteHeader(http.StatusOK)
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
		Info:    ipInfo,
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
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs)) // Serve "/" from staticDir
}

func (i *IpInfo) HandleEvent(event utils.Event) {
	nwChangedChannel <- event.Name
	// go utils.GetIpInfo(*i)
}

func WebServer(port string) {
	utils.RegisterListener([]string{"vpn-up", "vpn-down"}, &ipInfo)

	go utils.GetIpInfo(ipInfo)

	// Create a new Gorilla Mux router
	r := mux.NewRouter()

	// Config-related routes
	r.HandleFunc("/api/status", statusHandler).Methods("GET")
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
	fmt.Printf("Server starting on port %s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		utils.LogFatal(err)
	}
}
