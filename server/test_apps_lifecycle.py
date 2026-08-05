import importlib
import os
import subprocess
import tempfile
import unittest
from unittest.mock import call, patch

import core
from actions.apps import run_legacy_app_script, run_managed_apps_action, setup_apps_once
from modules import apps as apps_module
from utils.exec import run_command_logged


class AppLifecycleTests(unittest.TestCase):
    def setUp(self):
        self.testing = core.Testing
        self.app_script = core.AppScript
        core.Testing = False

    def tearDown(self):
        core.Testing = self.testing
        core.AppScript = self.app_script

    def test_managed_actions_run_directly_and_filter_disabled_apps(self):
        apps = [
            {
                "name": "enabled",
                "enabled": True,
                "setupCommands": ["install-enabled"],
                "upCommands": ["start-enabled"],
            },
            {
                "name": "disabled",
                "enabled": False,
                "setupCommands": ["install-disabled"],
                "upCommands": ["start-disabled"],
            },
        ]

        with patch("actions.apps.run_command_logged", return_value=None) as run:
            self.assertIsNone(run_managed_apps_action("setup", apps))
            self.assertIsNone(run_managed_apps_action("up", apps))

        self.assertEqual(
            run.call_args_list,
            [
                call(False, "/bin/sh", "-c", "install-enabled", log_command=False),
                call(False, "/bin/sh", "-c", "install-disabled", log_command=False),
                call(False, "/bin/sh", "-c", "start-enabled", log_command=False),
            ],
        )

    def test_managed_action_failure_does_not_expose_command(self):
        secret_command = "start-app --token secret-value"
        failure = subprocess.CalledProcessError(9, ["/bin/sh", "-c", secret_command])
        app = {
            "name": "example",
            "enabled": True,
            "upCommands": [secret_command],
        }

        with patch("actions.apps.run_command_logged", return_value=failure):
            err = run_managed_apps_action("up", [app])

        self.assertIn("example up", str(err))
        self.assertIn("exit status 9", str(err))
        self.assertNotIn("secret-value", str(err))

    def test_legacy_script_is_independent_and_optional(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            core.AppScript = os.path.join(temp_dir, "apps.sh")
            with patch("actions.apps.run_command_logged") as run:
                self.assertIsNone(run_legacy_app_script("up"))
                run.assert_not_called()

            with open(core.AppScript, "w") as script:
                script.write("#!/bin/sh\n")

            with patch("actions.apps.run_command_logged", return_value=None) as run:
                self.assertIsNone(run_legacy_app_script("down"))
                run.assert_called_once_with(False, core.AppScript, "down")

    def test_initializing_managed_apps_does_not_touch_apps_sh(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            core.AppScript = os.path.join(temp_dir, "apps.sh")
            original = "#!/bin/sh\n# user managed\n"
            with open(core.AppScript, "w") as script:
                script.write(original)

            with patch("modules.apps.register_module") as register:
                apps_module.init_apps_module()

            register.assert_called_once()
            with open(core.AppScript) as script:
                self.assertEqual(script.read(), original)

    def test_logged_command_reports_nonzero_exit(self):
        err = run_command_logged(False, "/bin/sh", "-c", "exit 7")

        self.assertIsInstance(err, subprocess.CalledProcessError)
        self.assertEqual(err.returncode, 7)

    def test_setup_runs_both_sources_once(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            marker = os.path.join(temp_dir, "initialized")
            with (
                patch("actions.apps.run_legacy_app_script", return_value=None) as legacy,
                patch("actions.apps.run_managed_apps_action", return_value=None) as managed,
            ):
                self.assertIsNone(setup_apps_once(marker))
                self.assertIsNone(setup_apps_once(marker))

            legacy.assert_called_once_with("setup")
            managed.assert_called_once_with("setup")
            self.assertTrue(os.path.exists(marker))

    def test_failed_setup_is_not_marked_complete(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            marker = os.path.join(temp_dir, "initialized")
            failure = subprocess.CalledProcessError(1, ["apps.sh", "setup"])
            with (
                patch("actions.apps.run_legacy_app_script", return_value=failure),
                patch("actions.apps.run_managed_apps_action", return_value=None),
            ):
                self.assertIs(setup_apps_once(marker), failure)

            self.assertFalse(os.path.exists(marker))

    def test_vpn_actions_dispatch_both_legacy_and_managed_hooks(self):
        vpn_up_module = importlib.import_module("actions.vpn_up")
        vpn_down_module = importlib.import_module("actions.vpn_down")
        spec = vpn_up_module.NetSpec(dev="tun0")

        with (
            patch("builtins.open"),
            patch.object(vpn_up_module, "run_command"),
            patch.object(vpn_up_module, "set_vpn_output_firewall"),
            patch.object(vpn_up_module, "publish_event"),
            patch.object(vpn_up_module, "run_legacy_app_script") as legacy_up,
            patch.object(vpn_up_module, "run_managed_apps_action") as managed_up,
        ):
            vpn_up_module.vpn_up(spec)

        legacy_up.assert_called_once_with("up")
        managed_up.assert_called_once_with("up")

        with (
            patch.object(vpn_down_module, "run_command"),
            patch.object(vpn_down_module, "restore_resolv_conf"),
            patch.object(vpn_down_module, "publish_event_sync"),
            patch.object(vpn_down_module, "run_legacy_app_script") as legacy_down,
            patch.object(vpn_down_module, "run_managed_apps_action") as managed_down,
        ):
            vpn_down_module.vpn_down()

        legacy_down.assert_called_once_with("down")
        managed_down.assert_called_once_with("down")


if __name__ == "__main__":
    unittest.main()
