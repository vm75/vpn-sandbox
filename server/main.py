import os
import sys
import signal
from utils import log_ln, log_f, log_fatal, init_log, backup_resolv_conf, smart_args, run_command_logged, add_signal_handler, signal_running, get_log_file
import core
from core import init_core, OpenVPNAction, WebServer, SHUTDOWN, VPN_UP, VPN_DOWN
from actions import save_openvpn_spec, vpn_up, vpn_down
from modules import init_openvpn, init_wireguard, init_proxy_module, HttpProxy, SocksProxy, openvpn_shutdown, wireguard_shutdown, init_apps_module
from webserver.webserver import start_webserver

def one_time_setup():
    marker_file = "/.initialized"
    backup_resolv_conf()
    
    if not os.path.exists(marker_file):
        if core.Testing:
            log_ln("Skipping apps setup for testing")
        elif os.path.exists(core.AppScript):
            log_f(f"Running one-time setup for apps script {core.AppScript}\n")
            try:
                run_command_logged(False, core.AppScript, "setup")
            except Exception as e:
                log_ln(str(e))
        else:
            log_ln("No apps script found")
        try:
            with open(marker_file, "w") as f:
                pass
        except OSError:
            pass

def main():
    ex = os.path.abspath(sys.argv[0])
    try:
        os.chdir(os.path.dirname(ex))
    except Exception as e:
        log_fatal(str(e))
        
    options, args = smart_args("--data|-d=/data:,--port|-p=80:,--test,--sudo", sys.argv[1:])
    data_dir = options["--data"].get_value()
    core.Testing = options["--test"].is_set()
    import utils
    utils.use_sudo = options["--sudo"].is_set()
    
    script_type = os.getenv("script_type", "")
    app_mode = WebServer
    if script_type and len(args) > 0 and args[0][:3] == "tun":
        app_mode = OpenVPNAction
        
    err = init_core(data_dir, app_mode)
    if err:
        log_fatal(str(err))
        
    if app_mode == OpenVPNAction:
        init_log(os.path.join(core.VarDir, f"vpn-{script_type}.log"))
        log_f(f"Running openvpn action {script_type}\n")
        if script_type == "up":
            log_ln("Saving openvpn spec file")
            save_openvpn_spec()
            log_ln("Signaling vpn up to main process")
            signal_running(core.ServerPidFile, VPN_UP)
        elif script_type == "down":
            if not core.Testing and os.path.exists(core.AppScript):
                log_ln("Stopping apps script synchronously")
                run_command_logged(False, core.AppScript, "down")
            log_ln("Signaling vpn down to main process")
            signal_running(core.ServerPidFile, VPN_DOWN)
        sys.exit(0)
        
    def handle_signal(sig):
        log_f(f"Received signal {sig}\n")
        if sig == VPN_UP:
            vpn_up(None)
        elif sig == VPN_DOWN:
            vpn_down()
        elif sig == SHUTDOWN:
            openvpn_shutdown()
            wireguard_shutdown()
            os._exit(0)
            
    add_signal_handler([VPN_UP, VPN_DOWN, SHUTDOWN], handle_signal)
    
    log_ln("Running one-time setup")
    one_time_setup()
    
    vpn_down()
    
    init_proxy_module(HttpProxy)
    init_proxy_module(SocksProxy)
    init_openvpn()
    init_wireguard()
    init_apps_module()
    
    start_webserver(options["--port"].get_value())

if __name__ == "__main__":
    main()
