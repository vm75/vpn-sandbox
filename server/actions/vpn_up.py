import os
import json
from utils import log_ln, log_error, run_command, publish_event, Event
import core
from .apps import run_legacy_app_script, run_managed_apps_action

class NetSpec:
    def __init__(self, dev="", domains=None, dns=None, vpn_gateway="", vpn_endpoint=""):
        self.dev = dev
        self.domains = domains or []
        self.dns = dns or []
        self.vpn_gateway = vpn_gateway
        self.vpn_endpoint = vpn_endpoint

    def to_dict(self):
        return {
            "dev": self.dev,
            "domains": self.domains,
            "dns": self.dns,
            "vpn_gateway": self.vpn_gateway,
            "vpn_endpoint": self.vpn_endpoint
        }
        
    @classmethod
    def from_dict(cls, d):
        return cls(
            dev=d.get("dev", ""),
            domains=d.get("domains", []),
            dns=d.get("dns", []),
            vpn_gateway=d.get("vpn_gateway", ""),
            vpn_endpoint=d.get("vpn_endpoint", "")
        )

def save_openvpn_spec():
    spec_file = os.path.join(core.VarDir, "openvpn.spec")
    
    net_spec = NetSpec(
        dev=os.getenv("dev", ""),
        vpn_gateway=os.getenv("route_vpn_gateway", ""),
        vpn_endpoint=os.getenv("trusted_ip", "")
    )
    
    i = 1
    while True:
        fopt = os.getenv(f"foreign_option_{i}")
        if not fopt:
            break
        if fopt.startswith("dhcp-option DOMAIN "):
            net_spec.domains.append(fopt[len("dhcp-option DOMAIN "):])
        elif fopt.startswith("dhcp-option DNS "):
            net_spec.dns.append(fopt[len("dhcp-option DNS "):])
        i += 1
        
    with open(spec_file, "w") as f:
        json.dump(net_spec.to_dict(), f, indent=2)

def retrieve_openvpn_spec():
    spec_file = os.path.join(core.VarDir, "openvpn.spec")
    try:
        with open(spec_file, "r") as f:
            data = json.load(f)
            return NetSpec.from_dict(data), None
    except Exception as e:
        return None, e

def set_vpn_output_firewall(command, dev, vpn_endpoint):
    run_command(False, command, "-F", "OUTPUT")
    run_command(False, command, "-A", "OUTPUT", "-o", "lo", "-j", "ACCEPT")
    run_command(False, command, "-A", "OUTPUT", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")
    run_command(False, command, "-A", "OUTPUT", "-o", dev, "-j", "ACCEPT")
    
    import ipaddress
    try:
        ip = ipaddress.ip_address(vpn_endpoint)
        is_ipv6_cmd = "ip6tables" in command
        if (is_ipv6_cmd and ip.version == 6) or (not is_ipv6_cmd and ip.version == 4):
            run_command(False, command, "-A", "OUTPUT", "-o", "eth0", "-d", vpn_endpoint, "-j", "ACCEPT")
    except ValueError:
        pass
        
    run_command(False, command, "-A", "OUTPUT", "-j", "DROP")

def vpn_up(net_spec=None):
    log_ln("VpnUp: Entry")
    
    if not net_spec:
        net_spec, err = retrieve_openvpn_spec()
        if not net_spec:
            log_error("No openvpn spec found", err)
            return
            
    sb = []
    if len(net_spec.domains) == 1:
        sb.append(f"domain {net_spec.domains[0]}\n")
    elif len(net_spec.domains) > 1:
        sb.append(f"search {' '.join(net_spec.domains)}\n")
        
    for nameserver in net_spec.dns:
        sb.append(f"nameserver {nameserver}\n")
        
    if core.Testing:
        publish_event(Event("vpn-up", {"dev": net_spec.dev}))
        log_ln("Skipping vpn up actions for testing")
        log_ln("resolv.conf: " + "".join(sb))
        return
        
    log_ln("Generating resolv.conf")
    try:
        with open("/etc/resolv.conf", "w") as f:
            f.write("".join(sb))
    except Exception as e:
        log_error("Error updating /etc/resolv.conf", e)
        
    log_ln("Setting routes")
    run_command(False, "/sbin/ip", "route", "del", "default")
    
    if not net_spec.vpn_gateway:
        run_command(False, "/sbin/ip", "route", "add", "default", "dev", net_spec.dev)
    else:
        run_command(False, "/sbin/ip", "route", "add", "default", "via", net_spec.vpn_gateway, "dev", net_spec.dev)
        
    if net_spec.vpn_endpoint:
        log_ln("host gateway: " + core.HostGateway)
        run_command(False, "/sbin/ip", "route", "add", net_spec.vpn_endpoint, "via", core.HostGateway)
        
    run_command(False, "/usr/sbin/iptables", "-F")
    run_command(False, "/usr/sbin/iptables", "-A", "INPUT", "-i", net_spec.dev, "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")
    run_command(False, "/usr/sbin/iptables", "-A", "INPUT", "-i", net_spec.dev, "-j", "DROP")
    
    set_vpn_output_firewall("/usr/sbin/iptables", net_spec.dev, net_spec.vpn_endpoint)
    set_vpn_output_firewall("/usr/sbin/ip6tables", net_spec.dev, net_spec.vpn_endpoint)
    
    log_ln("Triggering vpn-up actions")
    publish_event(Event("vpn-up", {"dev": net_spec.dev}))
    
    log_ln("Starting app lifecycle hooks")
    run_legacy_app_script("up")
    run_managed_apps_action("up")
