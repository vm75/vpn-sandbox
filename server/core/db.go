package core

import (
	"database/sql"
	"encoding/json"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB = nil
var createConfigsQuery = `CREATE TABLE IF NOT EXISTS configs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	config JSON NOT NULL
);`
var getConfigQuery = `SELECT config
	FROM configs
	WHERE name = ?;`
var saveConfigQuery = `INSERT OR REPLACE INTO configs
	(name, config)
	VALUES (?, ?);`

func initDb() error {
	var dbPath = filepath.Join(ConfigDir, "vpn-sandbox.db")
	var err error
	Db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	_, err = Db.Exec(createConfigsQuery)
	return err
}

func SaveConfig(name string, config map[string]interface{}) error {
	configStr, _ := json.Marshal(config)
	_, err := Db.Exec(saveConfigQuery, name, configStr)
	return err
}

func GetConfig(name string) (map[string]interface{}, error) {
	var configStr []byte
	row := Db.QueryRow(getConfigQuery, name)
	err := row.Scan(&configStr)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	json.Unmarshal(configStr, &config)
	return config, nil
}
