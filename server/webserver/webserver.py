import os
import json
import threading
import time
from datetime import datetime
from flask import Flask, request, jsonify, Response, send_from_directory
import core
from core import get_modules, enable_module, disable_module, restart_module, get_module_config, save_module_config, get_global_config, save_global_config, Version
from utils import get_ip_info, log_ln, EventListener, register_listener

app = Flask(__name__, static_folder="../static", static_url_path="/")

class StatusNotifier:
    def __init__(self):
        self.subscribers = []
        self.mu = threading.Lock()
        
    def subscribe(self):
        q = []
        with self.mu:
            self.subscribers.append(q)

        def unsubscribe():
            with self.mu:
                if q in self.subscribers:
                    self.subscribers.remove(q)

        return q, unsubscribe
        
    def notify(self, event):
        with self.mu:
            for q in list(self.subscribers):
                q.append(event)

status_updates = StatusNotifier()
status_heartbeat_interval = 15

class IpInfoCache:
    def __init__(self):
        self.mu = threading.Lock()
        self.refresh_mu = threading.Lock()
        self.output = {}
        self.executed_at = None
        self.last_event = ""
        self.last_event_at = None
        
    def mark_event(self, name, at):
        with self.mu:
            self.last_event = name
            self.last_event_at = at
            
    def store(self, output, executed_at):
        with self.mu:
            self.output = output
            self.executed_at = executed_at
            
    def snapshot(self):
        with self.mu:
            stale = True
            if self.executed_at and self.last_event_at:
                stale = self.executed_at < self.last_event_at
            elif not self.executed_at:
                stale = True
            else:
                stale = False
                
            return {
                "output": dict(self.output),
                "executedAt": self.executed_at.isoformat() + "Z" if self.executed_at else None,
                "event": self.last_event,
                "eventAt": self.last_event_at.isoformat() + "Z" if self.last_event_at else datetime.now().isoformat() + "Z",
                "stale": stale
            }
            
    def refresh(self):
        with self.refresh_mu:
            updated = {}
            err = get_ip_info(updated)
            if err:
                return False
            self.store(updated, datetime.now())
            return True
            
    def handle_event(self, event):
        self.mark_event(event.name, datetime.now())
        notify_status(event.name)
        if self.refresh():
            notify_status("ip-info")

ip_info = IpInfoCache()

def notify_status(event):
    status_updates.notify(event)

class StatusEventListener(EventListener):
    def handle_event(self, event):
        notify_status(event.name)

def get_status():
    global_cfg, _ = get_global_config()
    status = {
        "global": {
            "config": global_cfg
        }
    }
    
    for name, module in get_modules().items():
        status[name] = {
            "running": module.is_running(),
            "config": module.get_config({})
        }
        
    status["ipInfo"] = ip_info.snapshot()
    return status

@app.route("/api/status")
def status_handler():
    def stream():
        q, unsubscribe = status_updates.subscribe()
        try:
            data = json.dumps(get_status())
            yield f"data: {data}\n\n"
            
            while True:
                if q:
                    event = q.pop(0)
                    log_ln(f"Received event: {event}")
                    data = json.dumps(get_status())
                    yield f"data: {data}\n\n"
                else:
                    yield ": keepalive\n\n"
                time.sleep(1) # Simple polling for SSE for KISS
        finally:
            unsubscribe()
            
    return Response(stream(), content_type="text/event-stream")

@app.route("/api/force-refresh")
def force_refresh_handler():
    ip_info.mark_event("force", datetime.now())
    if not ip_info.refresh():
        notify_status("force")
        return "failed to refresh IP info", 502
    notify_status("ip-info")
    return jsonify(ip_info.snapshot())

@app.route("/api/version")
def version_handler():
    return jsonify({"version": Version})

@app.route("/api/config")
def get_global_config_handler():
    cfg, err = get_global_config()
    if err:
        return str(err), 500
    return jsonify(cfg)

@app.route("/api/config/save", methods=["POST"])
def save_global_config_handler():
    config = request.json
    err = save_global_config(config)
    if err:
        return str(err), 500
    return "", 200

@app.route("/api/<module>/status")
def get_module_status_handler(module):
    mod = get_modules().get(module)
    if not mod:
        return "module not found", 404
    is_running = mod.is_running()
    return jsonify({
        "running": is_running,
        "info": ip_info.snapshot()["output"]
    })

@app.route("/api/<module>/enable", methods=["POST"])
def enable_module_handler(module):
    start_now = request.args.get("start") == "true"
    err = enable_module(module, start_now)
    if err:
        return str(err), 500
    return "", 200

@app.route("/api/<module>/disable", methods=["POST"])
def disable_module_handler(module):
    stop_now = request.args.get("stop") == "true"
    err = disable_module(module, stop_now)
    if err:
        return str(err), 500
    return "", 200

@app.route("/api/<module>/restart", methods=["POST"])
def restart_module_handler(module):
    err = restart_module(module)
    if err:
        return str(err), 500
    return "", 200

@app.route("/api/<module>/config")
def get_module_config_handler(module):
    params = request.args.to_dict()
    cfg, err = get_module_config(module, params)
    if err:
        return str(err), 500
    return jsonify(cfg)

@app.route("/api/<module>/config/save", methods=["POST"])
def save_module_config_handler(module):
    params = request.args.to_dict()
    config = request.json
    err = save_module_config(module, params, config)
    if err:
        return str(err), 500
    return "", 200

def sanitize_path(input_path):
    import posixpath
    clean_path = posixpath.normpath("/" + input_path).lstrip("/")
    return os.path.join(core.VarDir, clean_path)

@app.route("/api/files")
def list_files_handler():
    rel_path = request.args.get("path", "")
    abs_path = sanitize_path(rel_path)
    
    if not os.path.exists(abs_path):
        return str(Exception("not found")), 500
        
    file_infos = []
    try:
        for f in os.listdir(abs_path):
            if f.endswith(".auth"):
                continue
            f_path = os.path.join(abs_path, f)
            file_infos.append({
                "name": f,
                "path": os.path.join(rel_path, f).lstrip("/"),
                "isDir": os.path.isdir(f_path)
            })
        file_infos.sort(key=lambda x: x["name"])
    except Exception as e:
        return str(e), 500
        
    return jsonify(file_infos)

@app.route("/api/file")
def file_content_handler():
    rel_path = request.args.get("path", "")
    abs_path = sanitize_path(rel_path)
    
    if not abs_path.startswith(core.VarDir) or abs_path.endswith(".auth"):
        return "Access denied", 403
        
    try:
        with open(abs_path, "rb") as f:
            content = f.read()
        return Response(content, content_type="text/plain")
    except Exception as e:
        return str(e), 500

@app.route("/", defaults={'path': ''})
@app.route("/<path:path>")
def serve_static(path):
    if path != "" and os.path.exists(os.path.join(app.static_folder, path)):
        return send_from_directory(app.static_folder, path)
    else:
        return send_from_directory(app.static_folder, 'index.html')

def start_webserver(port):
    register_listener(["vpn-up", "vpn-down"], ip_info)
    register_listener(["proxy-up", "proxy-down"], StatusEventListener())
    register_listener(["apps-changed"], StatusEventListener())
    
    ip_info.mark_event("startup", datetime.now())
    threading.Thread(target=lambda: notify_status("ip-info") if ip_info.refresh() else None).start()
    
    for _, module in get_modules().items():
        module.register_routes(app)
        
    log_ln(f"Server starting on port {port}")
    # Run flask with threaded=True which is default, good for SSE
    app.run(host="0.0.0.0", port=int(port), debug=False, use_reloader=False)
