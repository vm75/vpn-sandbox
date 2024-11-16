package core

import (
	"fmt"
	"sync"

	"github.com/gorilla/mux"
)

type Module interface {
	RegisterRoutes(r *mux.Router)
	IsRunning() bool
	Enable(startNow bool) error
	Disable(stopNow bool) error
	Restart() error
	GetConfig(params map[string]string) (map[string]interface{}, error)
	SaveConfig(params map[string]string, config map[string]interface{}) error
}

var modules = make(map[string]Module)
var modulesMutex sync.RWMutex

func RegisterModule(name string, module Module) {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()
	modules[name] = module
}

func GetModules() map[string]Module {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	var moduleList = map[string]Module{}
	for name, module := range modules {
		moduleList[name] = module
	}
	return moduleList
}

func GetModule(name string) Module {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	return modules[name]
}

func GetModuleStatus(name string) (bool, error) {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	if module, exists := modules[name]; exists {
		return module.IsRunning(), nil
	}
	return false, fmt.Errorf("module %s not found", name)
}

func EnableModule(name string, startNow bool) error {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	if module, exists := modules[name]; exists {
		return module.Enable(startNow)
	}
	return fmt.Errorf("module %s not found", name)
}

func DisableModule(name string, stopNow bool) error {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	if module, exists := modules[name]; exists {
		return module.Disable(stopNow)
	}
	return fmt.Errorf("module %s not found", name)
}

func RestartModule(name string) error {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	if module, exists := modules[name]; exists {
		return module.Restart()
	}
	return fmt.Errorf("module %s not found", name)
}

func GetModuleConfig(name string, params map[string]string) (map[string]interface{}, error) {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	if module, exists := modules[name]; exists {
		return module.GetConfig(params)
	}
	return nil, fmt.Errorf("module %s not found", name)
}

func SaveModuleConfig(name string, params map[string]string, config map[string]interface{}) error {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	if module, exists := modules[name]; exists {
		module.SaveConfig(params, config)
		return nil
	}

	return fmt.Errorf("module %s not found", name)
}
