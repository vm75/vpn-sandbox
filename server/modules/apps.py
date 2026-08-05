import os
import stat
import threading
import core
from core import Module, get_apps, get_app, save_app, delete_app, register_module
from utils import log_ln, log_error, publish_event, Event

_apps_mutex = threading.Lock()


def _default_app():
    return {
        "name": "",
        "enabled": True,
        "setupCommands": [],
        "upCommands": [],
        "downCommands": [],
    }


def _generate_apps_sh(apps):
    """Generate the apps.sh shell script from the given list of app configs."""
    lines = ["#!/bin/sh", "", "case \"$1\" in"]

    # setup case — all apps
    lines.append("  setup)")
    for app in apps:
        for cmd in app.get("setupCommands", []):
            cmd = cmd.strip()
            if cmd:
                lines.append(f"    {cmd}")
    lines.append("    ;;")

    # up case — enabled apps only
    lines.append("  up)")
    for app in apps:
        if not app.get("enabled", True):
            continue
        for cmd in app.get("upCommands", []):
            cmd = cmd.strip()
            if cmd:
                lines.append(f"    {cmd}")
    lines.append("    ;;")

    # down case — enabled apps only
    lines.append("  down)")
    for app in apps:
        if not app.get("enabled", True):
            continue
        for cmd in app.get("downCommands", []):
            cmd = cmd.strip()
            if cmd:
                lines.append(f"    {cmd}")
    lines.append("    ;;")

    lines.append("esac")
    lines.append("")
    return "\n".join(lines)


def regenerate_apps_sh():
    """Rebuild core.AppScript from all DB-stored apps."""
    apps, err = get_apps()
    if err:
        log_error("Error loading apps from DB", err)
        return err

    content = _generate_apps_sh(apps)
    try:
        with open(core.AppScript, "w") as f:
            f.write(content)
        os.chmod(core.AppScript, stat.S_IRWXU | stat.S_IRGRP | stat.S_IXGRP | stat.S_IROTH | stat.S_IXOTH)
        log_ln(f"Regenerated {core.AppScript} with {len(apps)} app(s)")
    except Exception as e:
        log_error("Error writing apps.sh", e)
        return e
    return None


class AppsModule(Module):
    """Module that manages user-defined apps stored in the database."""

    def __init__(self):
        self.name = "apps"

    def is_running(self):
        return False

    def get_config(self, params):
        apps, _ = get_apps()
        return {"apps": apps}

    def save_config(self, params, config):
        pass

    def register_routes(self, flask_app):
        from flask import request, jsonify

        @flask_app.route("/api/apps")
        def list_apps_handler():
            with _apps_mutex:
                apps, err = get_apps()
            if err:
                return str(err), 500
            return jsonify(apps)

        @flask_app.route("/api/apps/<string:name>")
        def get_app_handler(name):
            with _apps_mutex:
                app, err = get_app(name)
            if err:
                return str(err), 404
            return jsonify(app)

        @flask_app.route("/api/apps/save", methods=["POST"])
        def save_app_handler():
            config = request.json
            if not config or not config.get("name", "").strip():
                return "name is required", 400
            name = config["name"].strip()
            # Normalise — ensure all expected keys exist
            app = _default_app()
            app.update({
                "name": name,
                "enabled": bool(config.get("enabled", True)),
                "setupCommands": config.get("setupCommands", []),
                "upCommands": config.get("upCommands", []),
                "downCommands": config.get("downCommands", []),
            })
            with _apps_mutex:
                save_app(name, app)
                err = regenerate_apps_sh()
            if err:
                return str(err), 500
            publish_event(Event("apps-changed", {"name": name}))
            return "", 200

        @flask_app.route("/api/apps/<string:name>/delete", methods=["POST"])
        def delete_app_handler(name):
            with _apps_mutex:
                deleted = delete_app(name)
                if not deleted:
                    return "app not found", 404
                err = regenerate_apps_sh()
            if err:
                return str(err), 500
            publish_event(Event("apps-changed", {"name": name}))
            return "", 200

        @flask_app.route("/api/apps/<string:name>/enable", methods=["POST"])
        def enable_app_handler(name):
            return _set_app_enabled(name, True)

        @flask_app.route("/api/apps/<string:name>/disable", methods=["POST"])
        def disable_app_handler(name):
            return _set_app_enabled(name, False)

        def _set_app_enabled(name, enabled):
            with _apps_mutex:
                app, err = get_app(name)
                if err:
                    return str(err), 404
                app["enabled"] = enabled
                save_app(name, app)
                err = regenerate_apps_sh()
            if err:
                return str(err), 500
            publish_event(Event("apps-changed", {"name": name}))
            return "", 200


_module_instance = None


def init_apps_module():
    global _module_instance
    _module_instance = AppsModule()
    register_module("apps", _module_instance)

    # If there are DB-managed apps, regenerate apps.sh now. If DB is empty and
    # a hand-written apps.sh already exists on disk, leave it intact so existing
    # setups are not broken on first upgrade.
    apps, _ = get_apps()
    if apps:
        regenerate_apps_sh()
    else:
        log_ln("No DB apps found; leaving existing apps.sh (if any) unchanged")
