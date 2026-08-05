import os
import threading

import core
from utils import log_error, log_ln, run_command_logged


_apps_action_mutex = threading.Lock()
_COMMAND_KEYS = {
    "setup": "setupCommands",
    "up": "upCommands",
    "down": "downCommands",
}


def run_legacy_app_script(action):
    """Run the optional user-managed apps.sh hook."""
    if not os.path.isfile(core.AppScript):
        log_ln(f"No apps script found for {action}")
        return None

    err = run_command_logged(False, core.AppScript, action)
    if err:
        log_error(f"Error running apps script {action}", err)
    return err


def run_managed_apps_action(action, apps=None):
    """Run one lifecycle phase directly from the configured app records."""
    command_key = _COMMAND_KEYS.get(action)
    if command_key is None:
        return ValueError(f"unknown app action: {action}")

    if core.Testing:
        log_ln(f"Skipping managed apps {action} for testing")
        return None

    if apps is None:
        apps, err = core.get_apps()
        if err:
            log_error("Error loading managed apps", err)
            return err

    first_error = None
    with _apps_action_mutex:
        for app in apps:
            if action != "setup" and not app.get("enabled", True):
                continue

            name = app.get("name", "")
            for command in app.get(command_key, []):
                command = command.strip()
                if not command:
                    continue

                log_ln(f"Running managed app {name} {action}")
                err = run_command_logged(
                    False, "/bin/sh", "-c", command, log_command=False
                )
                if err:
                    return_code = getattr(err, "returncode", None)
                    detail = f" with exit status {return_code}" if return_code is not None else ""
                    action_err = RuntimeError(
                        f"managed app {name} {action} command failed{detail}"
                    )
                    log_error("Error running managed app command", action_err)
                    if first_error is None:
                        first_error = action_err

    return first_error


def setup_apps_once(marker_file="/.initialized"):
    """Run both setup sources once for this container instance."""
    if os.path.exists(marker_file):
        return None

    if core.Testing:
        log_ln("Skipping apps setup for testing")
    else:
        log_ln("Running one-time app setup")
        legacy_err = run_legacy_app_script("setup")
        managed_err = run_managed_apps_action("setup")
        if legacy_err or managed_err:
            err = legacy_err or managed_err
            log_ln("App setup failed; initialization will be retried on restart")
            return err

    try:
        with open(marker_file, "w"):
            pass
    except OSError as err:
        log_error("Error recording app initialization", err)
        return err
    return None
