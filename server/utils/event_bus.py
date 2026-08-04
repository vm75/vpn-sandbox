import threading

class Event:
    def __init__(self, name, context):
        self.name = name
        self.context = context

class EventListener:
    def handle_event(self, event: Event):
        pass

event_mutex = threading.Lock()
event_listeners = {}

def register_listener(event_names, listener: EventListener):
    with event_mutex:
        for name in event_names:
            if name not in event_listeners:
                event_listeners[name] = []
            event_listeners[name].append(listener)

def publish_event(event: Event):
    with event_mutex:
        listeners = event_listeners.get(event.name, [])[:]
    for listener in listeners:
        threading.Thread(target=listener.handle_event, args=(event,)).start()
