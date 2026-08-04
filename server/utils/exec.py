import subprocess
import os
import signal
import sys
from .log import log_ln, get_log_file

use_sudo = False

def run_command(is_elevated, command, *args):
    cmd_args = list(args)
    if is_elevated or use_sudo:
        cmd_args.insert(0, command)
        command = "sudo"
    
    log_ln(f"Running: {command} {' '.join(cmd_args)}")
    
    try:
        result = subprocess.run(
            [command] + cmd_args,
            stderr=get_log_file() if get_log_file() != sys.stdout else subprocess.STDOUT,
            stdout=subprocess.PIPE,
            preexec_fn=os.setsid,
            text=True
        )
        return result.stdout, None
    except Exception as e:
        return "", e

def run_command_logged(is_elevated, command, *args):
    cmd_args = list(args)
    if is_elevated or use_sudo:
        cmd_args.insert(0, command)
        command = "sudo"
    
    log_ln(f"Running: {command} {' '.join(cmd_args)}")
    
    try:
        subprocess.run(
            [command] + cmd_args,
            stdout=get_log_file(),
            stderr=get_log_file(),
            preexec_fn=os.setsid
        )
        return None
    except Exception as e:
        return e

def start_command(is_elevated, command, *args):
    cmd_args = list(args)
    if is_elevated or use_sudo:
        cmd_args.insert(0, command)
        command = "sudo"
    
    log_ln(f"Starting: {command} {' '.join(cmd_args)}")
    
    try:
        process = subprocess.Popen(
            [command] + cmd_args,
            stdout=get_log_file(),
            stderr=get_log_file(),
            preexec_fn=os.setsid
        )
        return process, None
    except Exception as e:
        return None, e

def is_running(process):
    return process is not None and process.poll() is None

def signal_cmd(process, sig):
    if is_running(process):
        try:
            process.send_signal(sig)
        except ProcessLookupError:
            pass

def signal_process(pid, sig):
    try:
        if use_sudo:
            run_command(True, "/bin/kill", f"-{sig}", str(pid))
        else:
            os.kill(pid, sig)
        return None
    except Exception as e:
        return e

def signal_running(pid_file, sig):
    if not os.path.exists(pid_file):
        log_ln(f"PID file {pid_file} not found")
        return False
    
    try:
        with open(pid_file, "r") as f:
            pid = int(f.read().strip())
    except Exception as e:
        log_ln(f"Error reading PID from file {pid_file}")
        return False
    
    err = signal_process(pid, 0)
    if err is not None:
        log_ln(f"Process with PID {pid} is not running")
        
    log_ln(f"Sending signal {sig} to process with PID {pid}")
    return signal_process(pid, sig) is None
