from utils import log_ln, run_command, publish_event_sync, Event, restore_resolv_conf
import core
from .apps import run_legacy_app_script, run_managed_apps_action

def vpn_down():
    log_ln("VpnDown: Entry")
    
    if not core.Testing:
        log_ln("Stopping app lifecycle hooks")
        run_legacy_app_script("down")
        run_managed_apps_action("down")
            
    log_ln("Triggering vpn down actions")
    publish_event_sync(Event("vpn-down", {}))
    
    log_ln("Restoring resolv.conf")
    restore_resolv_conf()
    
    log_ln("host gateway: " + core.HostGateway)
    
    if core.Testing:
        log_ln("Skipping vpn down actions for testing")
        return
        
    log_ln("Restoring routes")
    run_command(False, "/sbin/ip", "route", "del", "default")
    run_command(False, "/sbin/ip", "route", "add", "default", "via", core.HostGateway, "dev", "eth0")
    
    run_command(False, "/usr/sbin/iptables", "-F")
    run_command(False, "/usr/sbin/ip6tables", "-F")
    
    run_command(False, "/usr/sbin/iptables", "-A", "INPUT", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")
    run_command(False, "/usr/sbin/iptables", "-A", "INPUT", "-p", "tcp", "--dport", "80", "-j", "ACCEPT")
    run_command(False, "/usr/sbin/iptables", "-A", "INPUT", "-j", "DROP")
