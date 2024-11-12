package wireguard

import (
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

type Server struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

var createServersQuery = `CREATE TABLE IF NOT EXISTS wireguardServers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	template TEXT NOT NULL
);`
var getServersQuery = `SELECT name, template
	FROM wireguardServers;`
var getServerQuery = `SELECT name, template
	FROM wireguardServers
	WHERE name = ?;`
var insertServerQuery = `INSERT OR REPLACE INTO wireguardServers
	(name, template)
	VALUES (?, ?);`
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
		utils.LogFatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var config Server
		err := rows.Scan(
			&config.Name,
			&config.Template)
		if err != nil {
			utils.LogFatal(err)
			return templates
		}
		templates = append(templates, config)
	}
	return templates
}

func getWireguardServer(name string) *Server {
	var config Server
	row := core.Db.QueryRow(getServerQuery, name)
	err := row.Scan(
		&config.Name,
		&config.Template)
	if err != nil {
		utils.LogFatal(err)
		return nil
	}
	return &config
}

func saveWireguardServer(serverConfig Server) error {
	_, err := core.Db.Exec(insertServerQuery,
		serverConfig.Name,
		serverConfig.Template)
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
