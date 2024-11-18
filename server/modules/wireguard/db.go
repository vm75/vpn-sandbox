package wireguard

import (
	"encoding/json"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

type Server struct {
	Name      string              `json:"name"`
	Template  string              `json:"template"`
	HasParams bool                `json:"hasParams"`
	Endpoints []map[string]string `json:"endpoints"`
}

var createServersQuery = `CREATE TABLE IF NOT EXISTS wireguardServers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	template TEXT NOT NULL,
	hasParams BOOLEAN NOT NULL,
	endpoints JSON NOT NULL
);`
var getServersQuery = `SELECT name, template, hasParams, endpoints
	FROM wireguardServers;`
var getServerQuery = `SELECT name, template, hasParams, endpoints
	FROM wireguardServers
	WHERE name = ?;`
var insertServerQuery = `INSERT OR REPLACE INTO wireguardServers
	(name, template, hasParams, endpoints)
	VALUES (?, ?, ?, ?);`
var deleteServerQuery = `DELETE FROM wireguardServers WHERE name = ?;`

func initDb() {
	_, err := core.Db.Exec(createServersQuery)
	if err != nil {
		utils.LogFatal(err)
	}
}

func getWireguardServers() []Server {
	var templates []Server = make([]Server, 0)
	rows, err := core.Db.Query(getServersQuery)
	if err != nil {
		utils.LogError("Failed to get Wireguard servers", err)
	}
	defer rows.Close()
	var endpointsStr []byte
	for rows.Next() {
		var config Server
		err := rows.Scan(
			&config.Name,
			&config.Template,
			&config.HasParams,
			&endpointsStr)
		if err != nil {
			utils.LogError("Error getting Wireguard servers", err)
			return templates
		}
		json.Unmarshal(endpointsStr, &config.Endpoints)
		templates = append(templates, config)
	}
	return templates
}

func getWireguardServer(name string) *Server {
	var config Server
	row := core.Db.QueryRow(getServerQuery, name)
	var endpointsStr []byte
	err := row.Scan(
		&config.Name,
		&config.Template,
		&config.HasParams,
		&endpointsStr)
	if err != nil {
		utils.LogError("Error getting Wireguard server", err)
		return nil
	}
	json.Unmarshal(endpointsStr, &config.Endpoints)
	return &config
}

func saveWireguardServer(serverConfig Server) error {
	// Remove empty endpoints
	var savedNames = make(map[string]bool)
	for i, endpoint := range serverConfig.Endpoints {
		if endpoint["name"] == "" || savedNames[endpoint["name"]] {
			serverConfig.Endpoints = append(serverConfig.Endpoints[:i], serverConfig.Endpoints[i+1:]...)
		} else {
			savedNames[endpoint["name"]] = true
		}
	}

	endpointsStr, err := json.Marshal(serverConfig.Endpoints)
	if err != nil {
		utils.LogError("Error saving Wireguard server", err)
		return err
	}

	_, err = core.Db.Exec(insertServerQuery,
		serverConfig.Name,
		serverConfig.Template,
		serverConfig.HasParams,
		endpointsStr)
	if err != nil {
		utils.LogError("Error saving Wireguard server", err)
	}
	return err
}

func DeleteServer(name string) error {
	_, err := core.Db.Exec(deleteServerQuery, name)
	if err != nil {
		utils.LogError("Error deleting Wireguard server", err)
	}
	return err
}
