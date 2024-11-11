package core

import (
	"github.com/gorilla/mux"
)

type DefaultModule struct {
	Name string

	Config map[string]interface{}
}

func (d *DefaultModule) LoadConfig() {
	var err error
	d.Config, err = GetConfig(d.Name)
	if err != nil || d.Config == nil {
		d.Config = map[string]interface{}{"enabled": false}
		SaveConfig(d.Name, d.Config)
	}
}

func (d *DefaultModule) RegisterRoutes(r *mux.Router) {}

func (d *DefaultModule) GetStatus() (ModuleStatus, error) {
	return ModuleStatus{}, nil
}

func (d *DefaultModule) Enable(startNow bool) error {
	d.Config["enabled"] = true
	return SaveConfig(d.Name, d.Config)
}

func (d *DefaultModule) Disable(stopNow bool) error {
	d.Config["enabled"] = false
	return SaveConfig(d.Name, d.Config)
}

func (d *DefaultModule) Start() error {
	return nil
}

func (d *DefaultModule) Stop() error {
	return nil
}

func (d *DefaultModule) Restart() error {
	return nil
}

func (d *DefaultModule) GetConfig(params map[string]string) (map[string]interface{}, error) {
	return d.Config, nil
}

func (d *DefaultModule) SaveConfig(params map[string]string, config map[string]interface{}) error {
	d.Config["enabled"] = config["enabled"] == true
	return SaveConfig(d.Name, d.Config)
}
