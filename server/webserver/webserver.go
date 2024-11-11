package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

type IpInfo map[string]interface{}

var staticDir = "./static"
var ipInfo = IpInfo{}

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

func getGlobalConfig(w http.ResponseWriter, r *http.Request) {
	config, err := core.GetGlobalConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(config)
}

func saveGlobalConfig(w http.ResponseWriter, r *http.Request) {
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
func getModuleStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	status, err := core.GetModuleStatus(module)
	status.Info = ipInfo
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(status)
}

// Enable Module
func enableModule(w http.ResponseWriter, r *http.Request) {
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
func disableModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	stopNow := r.URL.Query().Get("stop") == "true"

	err := core.DisableModule(module, stopNow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// Start Module
func startModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	err := core.StartModule(module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// Stop Module
func stopModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	err := core.StopModule(module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// Restart Module
func restartModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	module := vars["module"]

	err := core.RestartModule(module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func getModuleConfig(w http.ResponseWriter, r *http.Request) {
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
func saveModuleConfig(w http.ResponseWriter, r *http.Request) {
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

// Separate function to handle static files
func handleStaticFiles(r *mux.Router) {
	// Serve static files from /static and root (/)
	fs := http.FileServer(http.Dir(staticDir))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs)) // Serve "/" from staticDir
}

func WebServer(port string) {
	utils.AddSignalHandler([]os.Signal{core.VPN_UP, core.VPN_DOWN}, func(_ os.Signal) {
		go utils.GetIpInfo(ipInfo)
	})

	go utils.GetIpInfo(ipInfo)

	// Create a new Gorilla Mux router
	r := mux.NewRouter()

	// Config-related routes
	r.HandleFunc("/api/config", getGlobalConfig).Methods("GET")
	r.HandleFunc("/api/config/save", saveGlobalConfig).Methods("POST")

	// Module
	r.HandleFunc("/api/{module}/status", getModuleStatus).Methods("GET")
	r.HandleFunc("/api/{module}/enable", enableModule).Methods("POST")
	r.HandleFunc("/api/{module}/disable", disableModule).Methods("POST")
	r.HandleFunc("/api/{module}/start", startModule).Methods("POST")
	r.HandleFunc("/api/{module}/stop", stopModule).Methods("POST")
	r.HandleFunc("/api/{module}/restart", restartModule).Methods("POST")
	r.HandleFunc("/api/{module}/config", getModuleConfig).Methods("GET")
	r.HandleFunc("/api/{module}/config/save", saveModuleConfig).Methods("POST")

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
