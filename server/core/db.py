import json
import os
import threading
import sqlite3

_settings_path = ""
_lock = threading.Lock()
_cache = {
    "configs": {},
    "apps": {}
}

def _save_atomic():
    temp_path = _settings_path + ".tmp"
    with open(temp_path, "w") as f:
        json.dump(_cache, f, indent=2)
    os.rename(temp_path, _settings_path)

def _load():
    global _cache
    if os.path.exists(_settings_path):
        with open(_settings_path, "r") as f:
            try:
                _cache = json.load(f)
            except json.JSONDecodeError:
                pass
    if "configs" not in _cache:
        _cache["configs"] = {}
    if "apps" not in _cache:
        _cache["apps"] = {}

def get_db():
    return None

def init_db(config_dir):
    global _settings_path
    _settings_path = os.path.join(config_dir, "settings.json")
    
    with _lock:
        _load()
        
    old_db_path = os.path.join(config_dir, "vpn-sandbox.db")
    if os.path.exists(old_db_path):
        migrated_any = False
        try:
            conn = sqlite3.connect(old_db_path)
            cursor = conn.cursor()
            
            try:
                cursor.execute("SELECT name, config FROM configs")
                for name, config_str in cursor.fetchall():
                    if name not in _cache["configs"]:
                        try:
                            _cache["configs"][name] = json.loads(config_str)
                            migrated_any = True
                        except Exception:
                            pass
            except Exception:
                pass
                
            try:
                cursor.execute("SELECT name, config FROM apps")
                for name, config_str in cursor.fetchall():
                    if name not in _cache["apps"]:
                        try:
                            _cache["apps"][name] = json.loads(config_str)
                            migrated_any = True
                        except Exception:
                            pass
            except Exception:
                pass
                
            if "wireguardServers" not in _cache["configs"]:
                _cache["configs"]["wireguardServers"] = []
                
            try:
                cursor.execute("SELECT name, template, hasParams, endpoints FROM wireguardServers")
                for row in cursor.fetchall():
                    name = row[0]
                    exists = any(s.get("name") == name for s in _cache["configs"]["wireguardServers"])
                    if not exists:
                        _cache["configs"]["wireguardServers"].append({
                            "name": name,
                            "template": row[1],
                            "hasParams": bool(row[2]),
                            "endpoints": json.loads(row[3])
                        })
                        migrated_any = True
            except Exception:
                pass
                
            if "openvpnServers" not in _cache["configs"]:
                _cache["configs"]["openvpnServers"] = []
                
            try:
                cursor.execute("SELECT name, template, hasParams, username, password, endpoints FROM openvpnServers")
                for row in cursor.fetchall():
                    name = row[0]
                    exists = any(s.get("name") == name for s in _cache["configs"]["openvpnServers"])
                    if not exists:
                        _cache["configs"]["openvpnServers"].append({
                            "name": name,
                            "template": row[1],
                            "hasParams": bool(row[2]),
                            "username": row[3],
                            "password": row[4],
                            "endpoints": json.loads(row[5])
                        })
                        migrated_any = True
            except Exception:
                pass
                
            conn.close()
        except Exception:
            pass
            
        if migrated_any:
            with _lock:
                _save_atomic()
                
        try:
            os.rename(old_db_path, old_db_path + ".bak")
        except Exception:
            pass

def save_config(name, config):
    with _lock:
        _cache["configs"][name] = config
        _save_atomic()

def get_config(name):
    with _lock:
        if name in _cache["configs"]:
            return _cache["configs"][name], None
        return None, Exception("not found")

# Apps CRUD

def get_apps():
    with _lock:
        result = []
        for name in sorted(_cache["apps"].keys()):
            result.append(_cache["apps"][name])
        return result, None

def get_app(name):
    with _lock:
        if name in _cache["apps"]:
            return _cache["apps"][name], None
        return None, Exception("not found")

def save_app(name, config):
    with _lock:
        _cache["apps"][name] = config
        _save_atomic()

def delete_app(name):
    with _lock:
        if name in _cache["apps"]:
            del _cache["apps"][name]
            _save_atomic()
            return True
        return False
