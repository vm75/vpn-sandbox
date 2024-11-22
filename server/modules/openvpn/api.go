package openvpn

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// List all servers
func listServersHandler(w http.ResponseWriter, r *http.Request) {
	var servers = getOpenVPNServers()
	json.NewEncoder(w).Encode(servers)
}

// Get a single server
func getServerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	server := getOpenVPNServer(name)
	if server == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(server)
}

// Create or update a server
func saveServerHandler(w http.ResponseWriter, r *http.Request) {
	var svr Server
	_ = json.NewDecoder(r.Body).Decode(&svr)

	err := saveOpenVPNServer(svr)
	if err != nil {
		http.Error(w, "Failed to save server", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Delete a template
func deleteServerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	err := DeleteServer(name)
	if err != nil {
		http.Error(w, "Failed to delete template", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
