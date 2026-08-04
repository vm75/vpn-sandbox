import sys

log_file = None

def init_log(file_path):
    global log_file
    log_file = open(file_path, "a")

def log_ln(*items):
    msg = " ".join(map(str, items))
    if log_file:
        log_file.write(msg + "\n")
        log_file.flush()
    print(msg, flush=True)

def log_f(fmt, *items):
    msg = fmt % items
    if log_file:
        log_file.write(msg)
        log_file.flush()
    sys.stdout.write(msg)
    sys.stdout.flush()

def log_error(msg, err):
    log_ln(msg, str(err))

def log_fatal(*items):
    log_ln(*items)
    sys.exit(1)

def get_log_file():
    return log_file if log_file else sys.stdout
