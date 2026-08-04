import signal
import threading

signal_handlers = {}
sig_channel = []

def init_signals(signals_list):
    for sig in signals_list:
        signal_handlers[sig] = []
        signal.signal(sig, _global_handler)

def _global_handler(sig, frame):
    handlers = signal_handlers.get(sig, [])
    for handler in handlers:
        threading.Thread(target=handler, args=(sig,)).start()

def real_time_signal(num):
    # Python's signal module has signal.SIGRTMIN and signal.SIGRTMAX in 3.3+ on linux
    if hasattr(signal, 'SIGRTMIN'):
        if 0 <= num <= (signal.SIGRTMAX - signal.SIGRTMIN):
            return signal.SIGRTMIN + num
    return None

def add_signal_handler(signals_list, handler):
    for sig in signals_list:
        if sig not in signal_handlers:
            signal_handlers[sig] = []
            signal.signal(sig, _global_handler)
        signal_handlers[sig].append(handler)
