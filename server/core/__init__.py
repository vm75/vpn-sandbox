import os
import signal
from utils import init_signals, real_time_signal, init_log, get_host_gateway, publish_event, Event
from .db import init_db, get_config, save_config, get_db, get_apps, get_app, save_app, delete_app
from .module import Module, get_modules, get_module, register_module, enable_module, disable_module, restart_module, get_module_config, save_module_config, get_module_status

DataDir = ""
ConfigDir = ""
VarDir = ""
ServerPidFile = ""
AppScript = ""
Testing = False
HostGateway = ""

class GlobalSettings:
    def __init__(self):
        self.vpn_types = ["OpenVPN", "Wireguard"]
        self.vpn_type = "OpenVPN"
        self.subnets = []
        self.proxy_username = ""
        self.proxy_password = ""

    def to_dict(self):
        return {
            "vpnTypes": self.vpn_types,
            "vpnType": self.vpn_type,
            "subnets": self.subnets,
            "proxyUsername": self.proxy_username,
            "proxyPassword": self.proxy_password
        }

    def from_dict(self, d):
        self.vpn_types = d.get("vpnTypes", self.vpn_types)
        self.vpn_type = d.get("vpnType", self.vpn_type)
        self.subnets = d.get("subnets", self.subnets)
        self.proxy_username = d.get("proxyUsername", self.proxy_username)
        self.proxy_password = d.get("proxyPassword", self.proxy_password)

GlobalConfig = GlobalSettings()

# AppMode
WebServer = 1
OpenVPNAction = 2

SHUTDOWN = signal.SIGTERM
VPN_UP = real_time_signal(1)
VPN_DOWN = real_time_signal(2)

Version = "dev"

def init_core(data_dir, app_mode):
    global DataDir, ConfigDir, VarDir, ServerPidFile, AppScript, HostGateway
    init_signals([SHUTDOWN, VPN_UP, VPN_DOWN])
    
    DataDir = data_dir
    ConfigDir = os.path.join(data_dir, "config")
    AppScript = os.path.join(data_dir, "apps.sh")
    VarDir = os.path.join(data_dir, "var")
    ServerPidFile = os.path.join(VarDir, "vpn-sandbox.pid")
    
    os.makedirs(ConfigDir, exist_ok=True)
    os.makedirs(VarDir, exist_ok=True)
    
    if app_mode == OpenVPNAction:
        return None
        
    import glob
    for pattern in ["*.log*", "*.pid"]:
        for file in glob.glob(os.path.join(VarDir, pattern)):
            try:
                os.remove(file)
            except OSError:
                pass
                
    init_log(os.path.join(VarDir, "vpn-sandbox.log"))
    HostGateway = get_host_gateway()
    
    try:
        with open(ServerPidFile, "w") as f:
            f.write(str(os.getpid()))
    except Exception as e:
        return e
        
    init_db(ConfigDir)
    
    saved_config, err = get_config("global")
    if err is None and saved_config:
        GlobalConfig.from_dict(saved_config)
    else:
        save_config("global", GlobalConfig.to_dict())
        
    return None

def get_global_config():
    return GlobalConfig.to_dict(), None

def save_global_config(config):
    current = GlobalConfig.to_dict()
    if current == config:
        return None
    GlobalConfig.from_dict(config)
    save_config("global", GlobalConfig.to_dict())
    publish_event(Event("global-config-changed", config))
    return None

def is_vpn_up():
    from .module import get_module
    openvpn_mod = get_module("openvpn")
    wg_mod = get_module("wireguard")
    return (openvpn_mod and openvpn_mod.is_running()) or (wg_mod and wg_mod.is_running())

def get_vpn_device():
    from .module import get_module
    if GlobalConfig.vpn_type == "OpenVPN":
        mod = get_module("openvpn")
        if mod and mod.is_running():
            return "tun0"
    elif GlobalConfig.vpn_type == "Wireguard":
        mod = get_module("wireguard")
        if mod and mod.is_running():
            return "wg0"
    return ""
