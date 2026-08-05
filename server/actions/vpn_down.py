from utils import log_ln, log_error, run_command_logged, run_command, publish_event_sync, Event, restore_resolv_conf
import core

def vpn_down():
    log_ln("VpnDown: Entry")
    
    if not core.Testing:
        log_ln("Stopping apps script")
        err = run_command_logged(False, core.AppScript, "down")
        if err:
            log_error("Error stopping apps script", err)
            
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
