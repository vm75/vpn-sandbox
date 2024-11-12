package wireguard

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// List all servers
func listServers(w http.ResponseWriter, r *http.Request) {
	var servers = getWireguardServers()
	json.NewEncoder(w).Encode(servers)
}

// Get a single server
func getServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	server := getWireguardServer(name)
	if server == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(server)
}

// Create or update a server
func saveServer(w http.ResponseWriter, r *http.Request) {
	var svr Server
	_ = json.NewDecoder(r.Body).Decode(&svr)

	err := saveWireguardServer(svr)
	if err != nil {
		http.Error(w, "Failed to save server", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Delete a template
func deleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	err := DeleteServer(name)
	if err != nil {
		http.Error(w, "Failed to delete template", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
