package openvpn

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

type Server struct {
	Name      string              `json:"name"`
	Template  string              `json:"template"`
	Username  string              `json:"username"`
	Password  string              `json:"password"`
	Endpoints []map[string]string `json:"endpoints"`
}

var createServersQuery = `CREATE TABLE IF NOT EXISTS openvpnServers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	template TEXT NOT NULL,
	username TEXT NOT NULL,
	password TEXT NOT NULL,
	endpoints JSON NOT NULL
);`
var getServersQuery = `SELECT name, template, username, password, endpoints
	FROM openvpnServers;`
var getServerQuery = `SELECT name, template, username, password, endpoints
	FROM openvpnServers
	WHERE name = ?;`
var insertServerQuery = `INSERT OR REPLACE INTO openvpnServers
	(name, template, username, password, endpoints)
	VALUES (?, ?, ?, ?, ?);`
var deleteServerQuery = `DELETE FROM openvpnServers WHERE name = ?;`

func initDb() {
	_, err := core.Db.Exec(createServersQuery)
	if err != nil {
		utils.LogFatal(err)
	}
}

func saveOvpnConfig() error {
	var server = getOpenVPNServer(openvpnConfig.ServerName)

	if server == nil {
		return errors.New("server not found")
	}

	var ovpn = server.Template
	var endpoint map[string]string = nil

	for _, entry := range server.Endpoints {
		if entry["name"] == openvpnConfig.ServerEndpoint {
			endpoint = entry
			break
		}
	}

	if endpoint == nil {
		return errors.New("endpoint not found")
	}

	for key, value := range endpoint {
		ovpn = strings.ReplaceAll(ovpn, "{{"+key+"}}", value)
	}

	auth := fmt.Sprintf("%s\n%s\n", server.Username, server.Password)

	configUpdated, configErr := utils.UpdateContent(ovpn, configFile)
	if configErr != nil {
		return configErr
	}
	authUpdated, authErr := utils.UpdateContent(auth, authFile)
	if authErr != nil {
		return authErr
	}

	if utils.IsRunning(openvpnCmd) && (configUpdated || authUpdated) {
		utils.LogLn("Configuration updated, restarting OpenVPN")
		killOpenVPN()
		openvpnCmd.Wait()
		go runOpenVPN()
	}

	return nil
}

func getOpenVPNServers() []Server {
	var templates []Server = make([]Server, 0)
	rows, err := core.Db.Query(getServersQuery)
	if err != nil {
		utils.LogFatal(err)
	}
	defer rows.Close()
	var endpointsStr []byte
	for rows.Next() {
		var config Server
		err := rows.Scan(
			&config.Name,
			&config.Template,
			&config.Username,
			&config.Password,
			&endpointsStr)
		if err != nil {
			utils.LogFatal(err)
			return templates
		}
		json.Unmarshal(endpointsStr, &config.Endpoints)
		templates = append(templates, config)
	}
	return templates
}

func getOpenVPNServer(name string) *Server {
	var config Server
	row := core.Db.QueryRow(getServerQuery, name)
	var endpointsStr []byte
	err := row.Scan(
		&config.Name,
		&config.Template,
		&config.Username,
		&config.Password,
		&endpointsStr)
	if err != nil {
		utils.LogFatal(err)
		return nil
	}
	json.Unmarshal(endpointsStr, &config.Endpoints)
	return &config
}

func saveOpenVPNServer(serverConfig Server) error {
	// Remove empty endpoints
	var savedNames = make(map[string]bool)
	for i, endpoint := range serverConfig.Endpoints {
		if endpoint["name"] == "" || savedNames[endpoint["name"]] {
			serverConfig.Endpoints = append(serverConfig.Endpoints[:i], serverConfig.Endpoints[i+1:]...)
		}
		savedNames[endpoint["name"]] = true
	}

	endpointsStr, err := json.Marshal(serverConfig.Endpoints)
	if err != nil {
		utils.LogFatal(err)
		return err
	}
	_, err = core.Db.Exec(insertServerQuery,
		serverConfig.Name,
		serverConfig.Template,
		serverConfig.Username,
		serverConfig.Password,
		endpointsStr)
	if err != nil {
		utils.LogFatal(err)
	}
	return err
}

func DeleteServer(name string) error {
	_, err := core.Db.Exec(deleteServerQuery, name)
	if err != nil {
		utils.LogFatal(err)
	}
	return err
}
