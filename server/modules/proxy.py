import os
import time
import threading
import core
from core import Module, get_config, save_config, register_module, GlobalConfig, get_vpn_device, is_vpn_up
from utils import log_ln, log_f, log_error, get_ipv4_addr, start_command, is_running, publish_event, Event

HttpProxy = 1
SocksProxy = 2

class ProxyModule(Module):
    def __init__(self, name, proxy_type, display_name, proxy_cmd, config_file, pid_file):
        self.name = name
        self.proxy_type = proxy_type
        self.display_name = display_name
        self.proxy_cmd = proxy_cmd
        self.config_file = config_file
        self.pid_file = pid_file
        self.config = {"enabled": False}
        self.cmd_mutex = threading.Lock()
        self.cmd_object = None
        self.done = None
        
        cfg, err = get_config(self.name)
        if not err and cfg:
            self.config = cfg
        else:
            save_config(self.name, self.config)

    def is_running(self):
        with self.cmd_mutex:
            return is_running(self.cmd_object)

    def enable(self, start_now):
        self.config["enabled"] = True
        save_config(self.name, self.config)
        if is_vpn_up() and start_now:
            threading.Thread(target=start_proxy, args=(self,)).start()

    def disable(self, stop_now):
        self.config["enabled"] = False
        save_config(self.name, self.config)
        if stop_now:
            stop_proxy(self)

    def restart(self):
        stop_proxy(self)
        if self.config.get("enabled", False) and is_vpn_up():
            threading.Thread(target=start_proxy, args=(self,), daemon=True).start()

    def get_config(self, params):
        return self.config

    def save_config(self, params, config):
        self.config["enabled"] = config.get("enabled", False)
        save_config(self.name, self.config)

    def handle_event(self, event):
        if event.name in ("global-config-changed", "vpn-up"):
            stop_proxy(self)
            threading.Thread(target=start_proxy, args=(self,)).start()
        elif event.name == "vpn-down":
            stop_proxy(self)

def quote_config_value(value):
    if "\r" in value or "\n" in value:
        raise ValueError("3proxy config values cannot contain newlines")
    return '"' + value.replace('"', '""') + '"'

def render_proxy_config(proxy_type, listen_addr, bind_addr, log_file, username, password):
    if not listen_addr or not bind_addr:
        raise ValueError("listen or bind address not found")
        
    quoted_log = quote_config_value(log_file)
    lines = [
        f"log {quoted_log}",
        "timeouts 1 5 30 60 180 1800 15 60 15 5",
        "maxconn 100",
        f"internal {listen_addr}",
        f"external {bind_addr}"
    ]
    
    if username and password:
        credentials = quote_config_value(f"{username}:CL:{password}")
        lines.extend([f"users {credentials}", "auth strong", "allow *"])
    else:
        lines.append("auth none")
        
    if proxy_type == HttpProxy:
        lines.append("proxy -p3128")
    elif proxy_type == SocksProxy:
        lines.append("socks -p1080")
        
    return "\n".join(lines) + "\n"

def update_proxy_config(p: ProxyModule):
    vpn_dev = get_vpn_device()
    if not vpn_dev:
        raise ValueError("VPN device not found")
        
    log_file = p.config_file
    if log_file.endswith(".cfg"):
        log_file = log_file[:-4] + ".log"
        
    content = render_proxy_config(
        p.proxy_type,
        get_ipv4_addr("eth0", True),
        get_ipv4_addr(vpn_dev, True),
        log_file,
        GlobalConfig.proxy_username,
        GlobalConfig.proxy_password
    )
    
    with open(p.config_file, "w") as f:
        f.write(content)
        
    os.chmod(p.config_file, 0o600)

def start_proxy(p: ProxyModule):
    if not p.config.get("enabled", False) or not is_vpn_up():
        return
        
    with p.cmd_mutex:
        if is_running(p.cmd_object):
            log_f(f"{p.display_name} is already running\n")
            return
            
    try:
        update_proxy_config(p)
    except Exception as e:
        log_error("Error updating runtime config", e)
        return
        
    cmd_obj, err = start_command(False, p.proxy_cmd[0], *p.proxy_cmd[1:])
    if err or not cmd_obj:
        log_error(f"Error starting {p.display_name}", err)
        return
        
    with p.cmd_mutex:
        p.cmd_object = cmd_obj
        p.done = threading.Event()
        done_event = p.done
        
    publish_event(Event("proxy-up", {}))
    log_f(f"{p.display_name} started with pid {cmd_obj.pid}\n")
    try:
        with open(p.pid_file, "w") as f:
            f.write(str(cmd_obj.pid))
    except Exception as e:
        log_error(f"Error writing {p.display_name} PID file", e)
        
    status = cmd_obj.wait()
    
    with p.cmd_mutex:
        if p.cmd_object == cmd_obj:
            p.cmd_object = None
            p.done = None
    done_event.set()
    
    publish_event(Event("proxy-down", {}))
    try:
        os.remove(p.pid_file)
    except OSError:
        pass
        
    log_f(f"{p.display_name} exited with status: {status}\n")

def stop_proxy(p: ProxyModule):
    log_f(f"Stopping {p.display_name}\n")
    with p.cmd_mutex:
        cmd = p.cmd_object
        done = p.done
        if not is_running(cmd):
            return
            
    try:
        cmd.terminate()
    except Exception as e:
        log_error(f"Error stopping {p.display_name}", e)
        return
        
    if not done.wait(5.0):
        log_f(f"{p.display_name} did not stop after 5 seconds; killing it\n")
        try:
            cmd.kill()
        except Exception as e:
            log_error(f"Error killing {p.display_name}", e)
        done.wait()

def init_proxy_module(proxy_type):
    if proxy_type == HttpProxy:
        cfg_file = os.path.join(core.VarDir, "3proxy-http.cfg")
        module = ProxyModule(
            name="http_proxy",
            proxy_type=HttpProxy,
            display_name="HTTP Proxy",
            proxy_cmd=["/usr/bin/3proxy", cfg_file],
            config_file=cfg_file,
            pid_file=os.path.join(core.VarDir, "3proxy-http.pid")
        )
    else:
        cfg_file = os.path.join(core.VarDir, "3proxy-socks.cfg")
        module = ProxyModule(
            name="socks_proxy",
            proxy_type=SocksProxy,
            display_name="SOCKS Proxy",
            proxy_cmd=["/usr/bin/3proxy", cfg_file],
            config_file=cfg_file,
            pid_file=os.path.join(core.VarDir, "3proxy-socks.pid")
        )
        
    register_module(module.name, module)
    from utils import register_listener
    register_listener(["global-config-changed", "vpn-up", "vpn-down"], module)
