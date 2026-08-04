import os
import json
import socket
import core
from core import Module, get_config, save_config, register_module, get_db
from utils import log_ln, log_error, file_exists, run_command
from actions import vpn_up, vpn_down, NetSpec

ModuleName = "wireguard"

class WireguardConfig:
    def __init__(self):
        self.enabled = False
        self.server_name = ""
        self.server_endpoint = ""
        
    def to_dict(self):
        return {
            "enabled": self.enabled,
            "serverName": self.server_name,
            "serverEndpoint": self.server_endpoint
        }
        
    def from_dict(self, d):
        self.enabled = d.get("enabled", self.enabled)
        self.server_name = d.get("serverName", self.server_name)
        self.server_endpoint = d.get("serverEndpoint", self.server_endpoint)

wireguard_config = WireguardConfig()

def init_db():
    cursor = get_db().cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS wireguardServers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            template TEXT NOT NULL,
            hasParams BOOLEAN NOT NULL,
            endpoints JSON NOT NULL
        )
    """)
    get_db().commit()

def get_wireguard_servers():
    cursor = get_db().cursor()
    cursor.execute("SELECT name, template, hasParams, endpoints FROM wireguardServers")
    servers = []
    for row in cursor.fetchall():
        servers.append({
            "name": row[0],
            "template": row[1],
            "hasParams": bool(row[2]),
            "endpoints": json.loads(row[3])
        })
    return servers

def get_wireguard_server(name):
    cursor = get_db().cursor()
    cursor.execute("SELECT name, template, hasParams, endpoints FROM wireguardServers WHERE name = ?", (name,))
    row = cursor.fetchone()
    if row:
        return {
            "name": row[0],
            "template": row[1],
            "hasParams": bool(row[2]),
            "endpoints": json.loads(row[3])
        }
    return None

def save_wireguard_server(data):
    endpoints = []
    seen = set()
    for ep in data.get("endpoints", []):
        name = ep.get("name")
        if name and name not in seen:
            seen.add(name)
            endpoints.append(ep)
            
    cursor = get_db().cursor()
    cursor.execute("""
        INSERT OR REPLACE INTO wireguardServers (name, template, hasParams, endpoints)
        VALUES (?, ?, ?, ?)
    """, (data.get("name"), data.get("template"), bool(data.get("hasParams")), json.dumps(endpoints)))
    get_db().commit()
    return None

def delete_server(name):
    cursor = get_db().cursor()
    cursor.execute("DELETE FROM wireguardServers WHERE name = ?", (name,))
    get_db().commit()
    return None

def is_tunnel_up():
    if not file_exists("/sys/class/net/wg0"):
        return False
    out, err = run_command(False, "/usr/bin/wg", "show", "wg0")
    if err is None and "peer: " in out:
        return True
    return False

def find_value(context, key, default):
    for line in context.split("\n"):
        line = line.strip()
        if line.startswith(key):
            parts = line.split("=", 1)
            if len(parts) == 2:
                return parts[1].strip()
    return default

def get_address(endpoint):
    try:
        host = endpoint.split(":")[0]
        ip = socket.gethostbyname(host)
        return ip
    except Exception as e:
        log_ln(str(e))
        return ""

def tunnel_up():
    if is_tunnel_up() or not wireguard_config.enabled:
        return
        
    server = get_wireguard_server(wireguard_config.server_name)
    if not server:
        log_ln("server not found")
        return
        
    endpoint = None
    for ep in server["endpoints"]:
        if ep.get("name") == wireguard_config.server_endpoint:
            endpoint = ep
            break
            
    wg_config = server["template"]
    if endpoint:
        for k, v in endpoint.items():
            wg_config = wg_config.replace("{{" + k + "}}", str(v))
            
    private_key = find_value(wg_config, "PrivateKey", "")
    peer_public_key = find_value(wg_config, "PublicKey", "")
    ep_val = find_value(wg_config, "Endpoint", "")
    dns_val = find_value(wg_config, "DNS", "1.1.1.1, 1.0.0.1")
    vpn_address = get_address(ep_val)
    address = find_value(wg_config, "Address", vpn_address)
    allowed_ips = find_value(wg_config, "AllowedIPs", "0.0.0.0/0")
    
    run_command(False, "/sbin/ip", "link", "add", "dev", "wg0", "type", "wireguard")
    
    with open("/tmp/wg0.key", "w") as f:
        f.write(private_key)
        
    run_command(False, "/usr/bin/wg", "set", "wg0",
        "private-key", "/tmp/wg0.key",
        "peer", peer_public_key,
        "endpoint", ep_val,
        "allowed-ips", allowed_ips)
        
    try:
        os.remove("/tmp/wg0.key")
    except Exception:
        pass
        
    run_command(False, "/sbin/ip", "address", "add", address, "dev", "wg0")
    run_command(False, "/sbin/ip", "link", "set", "up", "dev", "wg0")
    
    if not is_tunnel_up():
        log_ln("Tunnel up failed")
        return
        
    import threading
    dns_list = dns_val.replace(",", " ").split()
    threading.Thread(target=vpn_up, args=(NetSpec(
        dev="wg0",
        dns=dns_list,
        vpn_endpoint=vpn_address
    ),)).start()

def tunnel_down():
    run_command(False, "/sbin/ip", "link", "set", "down", "dev", "wg0")
    run_command(False, "/sbin/ip", "link", "del", "dev", "wg0")
    if is_tunnel_up():
        log_ln("Tunnel down failed")
    vpn_down()

class WireguardModule(Module):
    def register_routes(self, app):
        from flask import request, jsonify
        
        @app.route("/api/wireguard/servers", methods=["GET"])
        def list_wg_servers():
            return jsonify(get_wireguard_servers())
            
        @app.route("/api/wireguard/servers/<name>", methods=["GET"])
        def get_wg_server(name):
            srv = get_wireguard_server(name)
            if not srv:
                return "Server not found", 404
            return jsonify(srv)
            
        @app.route("/api/wireguard/servers/save", methods=["POST"])
        def save_wg_server():
            err = save_wireguard_server(request.json)
            if err:
                return "Failed to save server", 500
            return "", 200
            
        @app.route("/api/wireguard/servers/delete/<name>", methods=["DELETE"])
        def del_wg_server(name):
            err = delete_server(name)
            if err:
                return "Failed to delete template", 500
            return "", 200

    def is_running(self):
        return is_tunnel_up()

    def enable(self, start_now):
        wireguard_config.enabled = True
        save_config(ModuleName, wireguard_config.to_dict())
        if start_now:
            import threading
            threading.Thread(target=tunnel_up).start()
            
    def disable(self, stop_now):
        wireguard_config.enabled = False
        save_config(ModuleName, wireguard_config.to_dict())
        if stop_now:
            tunnel_down()
            
    def restart(self):
        tunnel_down()
        tunnel_up()

    def get_config(self, params):
        cfg = wireguard_config.to_dict()
        cfg["servers"] = get_wireguard_servers()
        return cfg

    def save_config(self, params, config):
        current = wireguard_config.to_dict()
        if current == config:
            return
        wireguard_config.from_dict(config)
        save_config(ModuleName, wireguard_config.to_dict())
        
        tunnel_down()
        import threading
        threading.Thread(target=tunnel_up).start()

def shutdown():
    tunnel_down()

def init_module():
    init_db()
    saved_config, err = get_config(ModuleName)
    if not err and saved_config:
        wireguard_config.from_dict(saved_config)
    else:
        save_config(ModuleName, wireguard_config.to_dict())
        
    register_module(ModuleName, WireguardModule())
    
    if wireguard_config.enabled:
        import threading
        threading.Thread(target=tunnel_up).start()
