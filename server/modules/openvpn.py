import os
import time
import subprocess
import json
import core
from core import Module, get_config, save_config, register_module, get_db
from utils import log_ln, log_f, log_error, file_exists, start_command, is_running, run_command, update_content
from core import get_module

ModuleName = "openvpn"

class OpenVPNConfig:
    def __init__(self):
        self.enabled = False
        self.server_name = ""
        self.server_endpoint = ""
        self.log_level = 0
        self.retry_interval = 300
    
    def to_dict(self):
        return {
            "enabled": self.enabled,
            "serverName": self.server_name,
            "serverEndpoint": self.server_endpoint,
            "logLevel": self.log_level,
            "retryInterval": self.retry_interval
        }
        
    def from_dict(self, d):
        self.enabled = d.get("enabled", self.enabled)
        self.server_name = d.get("serverName", self.server_name)
        self.server_endpoint = d.get("serverEndpoint", self.server_endpoint)
        self.log_level = d.get("logLevel", self.log_level)
        self.retry_interval = d.get("retryInterval", self.retry_interval)

openvpn_config = OpenVPNConfig()
openvpn_cmd = None
config_file = ""
auth_file = ""
pid_file = ""
log_file = ""
status_file = ""
status_update_interval = 60

class Server:
    def __init__(self, name, template, has_params, username, password, endpoints):
        self.name = name
        self.template = template
        self.has_params = has_params
        self.username = username
        self.password = password
        self.endpoints = endpoints

def init_db():
    cursor = get_db().cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS openvpnServers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            template TEXT NOT NULL,
            hasParams BOOLEAN NOT NULL,
            username TEXT NOT NULL,
            password TEXT NOT NULL,
            endpoints JSON NOT NULL
        )
    """)
    get_db().commit()

def get_openvpn_servers():
    cursor = get_db().cursor()
    cursor.execute("SELECT name, template, hasParams, username, password, endpoints FROM openvpnServers")
    servers = []
    for row in cursor.fetchall():
        servers.append({
            "name": row[0],
            "template": row[1],
            "hasParams": bool(row[2]),
            "username": row[3],
            "password": row[4],
            "endpoints": json.loads(row[5])
        })
    return servers

def get_openvpn_server(name):
    cursor = get_db().cursor()
    cursor.execute("SELECT name, template, hasParams, username, password, endpoints FROM openvpnServers WHERE name = ?", (name,))
    row = cursor.fetchone()
    if row:
        return Server(row[0], row[1], bool(row[2]), row[3], row[4], json.loads(row[5]))
    return None

def save_openvpn_server(data):
    endpoints = []
    seen = set()
    for ep in data.get("endpoints", []):
        name = ep.get("name")
        if name and name not in seen:
            seen.add(name)
            endpoints.append(ep)
            
    cursor = get_db().cursor()
    cursor.execute("""
        INSERT OR REPLACE INTO openvpnServers (name, template, hasParams, username, password, endpoints)
        VALUES (?, ?, ?, ?, ?, ?)
    """, (data.get("name"), data.get("template"), bool(data.get("hasParams")), data.get("username", ""), data.get("password", ""), json.dumps(endpoints)))
    get_db().commit()
    return None

def delete_server(name):
    cursor = get_db().cursor()
    cursor.execute("DELETE FROM openvpnServers WHERE name = ?", (name,))
    get_db().commit()
    return None

def save_ovpn_config():
    global openvpn_cmd
    server = get_openvpn_server(openvpn_config.server_name)
    if not server:
        return Exception("server not found")
        
    ovpn = server.template
    endpoint = None
    for ep in server.endpoints:
        if ep.get("name") == openvpn_config.server_endpoint:
            endpoint = ep
            break
            
    if not endpoint:
        return Exception("endpoint not found")
        
    for k, v in endpoint.items():
        ovpn = ovpn.replace("{{" + k + "}}", str(v))
        
    auth = f"{server.username}\n{server.password}\n"
    
    config_updated, _ = update_content(ovpn, config_file)
    auth_updated, _ = update_content(auth, auth_file)
    
    if is_running(openvpn_cmd) and (config_updated or auth_updated):
        log_ln("Configuration updated, restarting OpenVPN")
        kill_openvpn()
        if openvpn_cmd:
            openvpn_cmd.wait()
        import threading
        threading.Thread(target=run_openvpn).start()
        
    return None

def kill_openvpn():
    run_command(False, "/usr/bin/pkill", "-15", "-x", "openvpn")

def run_openvpn():
    global openvpn_cmd
    if is_running(openvpn_cmd) or not openvpn_config.enabled:
        return
        
    save_ovpn_config()
    
    if not file_exists(config_file) or not file_exists(auth_file):
        log_ln("VPN config/auth file(s) not found")
        return
        
    import sys
    exec_path = sys.executable
    script_path = os.path.abspath(sys.argv[0])
    
    while openvpn_config.enabled:
        retry_interval = str(openvpn_config.retry_interval)
        log_ln("Starting OpenVPN")
        openvpn_cmd, err = start_command(False, 
            "openvpn",
            "--client",
            "--cd", core.VarDir,
            "--config", config_file,
            "--auth-user-pass", auth_file,
            "--auth-nocache",
            "--verb", str(openvpn_config.log_level),
            "--log", log_file,
            "--status", status_file, str(status_update_interval),
            "--ping-restart", retry_interval,
            "--connect-retry-max", "3",
            "--script-security", "2",
            "--setenv", "script_type", "up",
            "--up", f"{exec_path} {script_path} tun up", "--up-delay",
            "--setenv", "script_type", "down",
            "--down", f"{exec_path} {script_path} tun down",
            "--up-restart",
            "--pull-filter", "ignore", "route-ipv6",
            "--pull-filter", "ignore", "ifconfig-ipv6",
            "--pull-filter", "ignore", "block-outside-dns",
            "--redirect-gateway", "def1",
            "--remote-cert-tls", "server",
            "--data-ciphers", "AES-256-GCM:AES-128-GCM:CHACHA20-POLY1305:AES-256-CBC:AES-128-CBC",
            "--writepid", pid_file
        )
        
        if err:
            log_ln(str(err))
            sleep_for = max(openvpn_config.retry_interval, 60)
            time.sleep(sleep_for)
        else:
            if openvpn_cmd and openvpn_cmd.pid:
                log_ln(f"OpenVPN started with pid {openvpn_cmd.pid}")
            status = openvpn_cmd.wait() if openvpn_cmd else -1
            log_f(f"OpenVPN exited with status: {status}\n")
            sleep_for = max(openvpn_config.retry_interval, 60)
            time.sleep(sleep_for)
            
        if not openvpn_config.enabled:
            break

class OpenVPNModule(Module):
    def register_routes(self, app):
        from flask import request, jsonify
        
        @app.route("/api/openvpn/servers", methods=["GET"])
        def list_servers():
            return jsonify(get_openvpn_servers())
            
        @app.route("/api/openvpn/servers/<name>", methods=["GET"])
        def get_server(name):
            srv = get_openvpn_server(name)
            if not srv:
                return "Server not found", 404
            return jsonify(srv.__dict__)
            
        @app.route("/api/openvpn/servers/save", methods=["POST"])
        def save_server():
            err = save_openvpn_server(request.json)
            if err:
                return "Failed to save server", 500
            return "", 200
            
        @app.route("/api/openvpn/servers/delete/<name>", methods=["DELETE"])
        def del_server(name):
            err = delete_server(name)
            if err:
                return "Failed to delete template", 500
            return "", 200

    def is_running(self):
        if not file_exists("/sys/class/net/tun0"):
            return False
            
        out, err = run_command(False, "/sbin/ip", "a", "show", "dev", "tun0")
        if err or "state DOWN" in out:
            return False
            
        if not file_exists(status_file):
            return False
            
        try:
            with open(status_file, "r") as f:
                content = f.read()
            updated_time_str = ""
            for line in content.split("\n"):
                if line.startswith("Updated,"):
                    updated_time_str = line[len("Updated,"):]
                    break
            from datetime import datetime
            updated_time = datetime.strptime(updated_time_str, "%Y-%m-%d %H:%M:%S")
            from datetime import timedelta
            if datetime.now() - updated_time <= timedelta(seconds=status_update_interval):
                return True
        except Exception:
            pass
        return False

    def enable(self, start_now):
        openvpn_config.enabled = True
        save_config(ModuleName, openvpn_config.to_dict())
        if start_now:
            import threading
            threading.Thread(target=run_openvpn).start()
            
    def disable(self, stop_now):
        openvpn_config.enabled = False
        save_config(ModuleName, openvpn_config.to_dict())
        if stop_now:
            kill_openvpn()
            
    def restart(self):
        kill_openvpn()

    def get_config(self, params):
        cfg = openvpn_config.to_dict()
        cfg["servers"] = get_openvpn_servers()
        return cfg

    def save_config(self, params, config):
        current = openvpn_config.to_dict()
        if current == config:
            return
        openvpn_config.from_dict(config)
        save_config(ModuleName, openvpn_config.to_dict())
        save_ovpn_config()
        
        kill_openvpn()
        import threading
        threading.Thread(target=run_openvpn).start()

def shutdown():
    openvpn_config.enabled = False
    kill_openvpn()
    if is_running(openvpn_cmd):
        openvpn_cmd.wait()

def init_module():
    global config_file, auth_file, pid_file, log_file, status_file
    init_db()
    
    config_file = os.path.join(core.VarDir, "vpn.ovpn")
    auth_file = os.path.join(core.VarDir, "vpn.auth")
    pid_file = os.path.join(core.VarDir, "openvpn.pid")
    log_file = os.path.join(core.VarDir, "openvpn.log")
    status_file = os.path.join(core.VarDir, "openvpn.status")
    
    saved_config, err = get_config(ModuleName)
    if not err and saved_config:
        openvpn_config.from_dict(saved_config)
    else:
        save_config(ModuleName, openvpn_config.to_dict())
        
    register_module(ModuleName, OpenVPNModule())
    
    if openvpn_config.enabled:
        import threading
        threading.Thread(target=run_openvpn).start()
