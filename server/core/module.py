import threading

class Module:
    def register_routes(self, app):
        pass
    def is_running(self):
        return False
    def enable(self, start_now):
        pass
    def disable(self, stop_now):
        pass
    def restart(self):
        pass
    def get_config(self, params):
        return {}
    def save_config(self, params, config):
        pass

modules = {}
modules_mutex = threading.Lock()

def register_module(name, module):
    with modules_mutex:
        modules[name] = module

def get_modules():
    with modules_mutex:
        return dict(modules)

def get_module(name):
    with modules_mutex:
        return modules.get(name)

def get_module_status(name):
    with modules_mutex:
        module = modules.get(name)
        if module:
            return module.is_running(), None
    return False, Exception(f"module {name} not found")

def enable_module(name, start_now):
    with modules_mutex:
        module = modules.get(name)
    if module:
        try:
            module.enable(start_now)
            return None
        except Exception as e:
            return e
    return Exception(f"module {name} not found")

def disable_module(name, stop_now):
    with modules_mutex:
        module = modules.get(name)
    if module:
        try:
            module.disable(stop_now)
            return None
        except Exception as e:
            return e
    return Exception(f"module {name} not found")

def restart_module(name):
    with modules_mutex:
        module = modules.get(name)
    if module:
        try:
            module.restart()
            return None
        except Exception as e:
            return e
    return Exception(f"module {name} not found")

def get_module_config(name, params):
    with modules_mutex:
        module = modules.get(name)
    if module:
        try:
            return module.get_config(params), None
        except Exception as e:
            return None, e
    return None, Exception(f"module {name} not found")

def save_module_config(name, params, config):
    with modules_mutex:
        module = modules.get(name)
    if module:
        try:
            module.save_config(params, config)
            return None
        except Exception as e:
            return e
    return Exception(f"module {name} not found")
