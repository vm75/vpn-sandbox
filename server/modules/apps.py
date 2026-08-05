import threading
from core import Module, get_apps, get_app, save_app, delete_app, register_module, is_vpn_up
from actions import run_managed_apps_action
from utils import publish_event, Event

_apps_mutex = threading.Lock()


def _default_app():
    return {
        "name": "",
        "enabled": True,
        "setupCommands": [],
        "upCommands": [],
        "downCommands": [],
    }


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
                previous, previous_err = get_app(name)
                if previous_err is None and previous == app:
                    return "", 200

                vpn_is_up = is_vpn_up()
                previous_was_running = bool(
                    previous_err is None
                    and previous.get("enabled", True)
                    and vpn_is_up
                )
                if previous_was_running:
                    err = run_managed_apps_action("down", [previous])
                    if err:
                        return str(err), 500

                setup_changed = (
                    previous_err is not None
                    or previous.get("setupCommands", []) != app["setupCommands"]
                )
                if setup_changed:
                    err = run_managed_apps_action("setup", [app])
                    if err:
                        if previous_was_running:
                            run_managed_apps_action("up", [previous])
                        return str(err), 500

                save_app(name, app)
                if app["enabled"] and vpn_is_up:
                    err = run_managed_apps_action("up", [app])
                    if err:
                        publish_event(Event("apps-changed", {"name": name}))
                        return str(err), 500
            publish_event(Event("apps-changed", {"name": name}))
            return "", 200

        @flask_app.route("/api/apps/<string:name>/delete", methods=["POST"])
        def delete_app_handler(name):
            with _apps_mutex:
                app, err = get_app(name)
                if err:
                    return "app not found", 404
                if app.get("enabled", True) and is_vpn_up():
                    err = run_managed_apps_action("down", [app])
                    if err:
                        return str(err), 500
                deleted = delete_app(name)
                if not deleted:
                    return "app not found", 404
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
                if app.get("enabled", True) == enabled:
                    return "", 200
                if is_vpn_up():
                    action = "up" if enabled else "down"
                    action_app = dict(app)
                    action_app["enabled"] = True
                    err = run_managed_apps_action(action, [action_app])
                    if err:
                        return str(err), 500
                app["enabled"] = enabled
                save_app(name, app)
            publish_event(Event("apps-changed", {"name": name}))
            return "", 200


_module_instance = None


def init_apps_module():
    global _module_instance
    _module_instance = AppsModule()
    register_module("apps", _module_instance)
