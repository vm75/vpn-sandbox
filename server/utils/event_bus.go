package utils

import (
	"sync"
)

// Event represents an event with a name and context.
type Event struct {
	Name    string
	Context map[string]interface{}
}

// EventListener represents an event listener.
type EventListener interface {
	HandleEvent(Event)
}

var eventMutex = sync.RWMutex{}
var eventListeners = make(map[string][]EventListener)

// RegisterListener registers a new listener to the event bus.
func RegisterListener(eventName string, listener EventListener) {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	if _, exists := eventListeners[eventName]; !exists {
		eventListeners[eventName] = make([]EventListener, 0)
	}

	eventListeners[eventName] = append(eventListeners[eventName], listener)
}

// PublishEvent publishes an event with a given context to all registered listeners.
func PublishEvent(event Event) {
	eventMutex.RLock()
	defer eventMutex.RUnlock()

	if _, exists := eventListeners[event.Name]; !exists {
		return
	}

	for _, listener := range eventListeners[event.Name] {
		// Each listener runs in its own goroutine to prevent blocking.
		go listener.HandleEvent(event)
	}
}
