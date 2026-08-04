import os
import subprocess
import json
import re
from .exec import run_command
from .file import file_exists
from .log import log_ln, log_error

RESOLV_CONF_BACKUP = "/etc/resolv.conf.bak"

def get_default_gateway():
    try:
        output = subprocess.check_output(["ip", "r"], text=True)
        for line in output.split("\n"):
            if line.startswith("default via"):
                return line.split(" ")[2], None
    except Exception as e:
        return "", e
    return "", None

def backup_resolv_conf():
    if not file_exists(RESOLV_CONF_BACKUP):
        run_command(False, "/bin/cp", "/etc/resolv.conf", RESOLV_CONF_BACKUP)
        if not file_exists(RESOLV_CONF_BACKUP):
            log_error(f"Error creating {RESOLV_CONF_BACKUP}", Exception("cp failed"))

def restore_resolv_conf():
    if file_exists(RESOLV_CONF_BACKUP):
        with open(RESOLV_CONF_BACKUP, "r") as f:
            content = f.read()
        with open("/etc/resolv.conf", "w") as f:
            f.write(content)

def get_host_gateway():
    out, _ = run_command(False, "/sbin/ip", "route", "show", "default")
    if out:
        for line in out.split("\n"):
            if line.startswith("default via"):
                gw = line.split(" ")[2]
                if gw:
                    return gw
    
    if file_exists(RESOLV_CONF_BACKUP):
        with open(RESOLV_CONF_BACKUP, "r") as f:
            for line in f:
                if line.startswith("nameserver"):
                    return line.split(" ")[1].strip()
    return ""

def get_ip_info(ip_info_dict):
    log_ln("get ip info")
    try:
        import urllib.request
        import ssl
        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        req = urllib.request.Request("https://ipinfo.io/json")
        with urllib.request.urlopen(req, context=ctx, timeout=10) as response:
            out = response.read().decode('utf-8')
        ip_info_dict.clear()
        ip_info_dict.update(json.loads(out))
        return None
    except Exception as e:
        log_error("Error getting ip info", e)
        return e

def get_ipv4_addr(dev, strip_mask=False):
    try:
        out = subprocess.check_output(["/sbin/ip", "a", "s", dev], text=True)
        match = re.search(r'([0-9]{1,3}(?:\.[0-9]{1,3}){3})\/[0-9]{1,2}', out)
        if match:
            addr = match.group(0)
            if strip_mask:
                addr = addr.split("/")[0]
            return addr
    except Exception:
        pass
    return ""
