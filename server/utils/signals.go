package utils

import (
	"os"
	"os/signal"
	"syscall"
)

type SignalHandler func(os.Signal)

const (
	SIGRTMIN = 34
	SIGRTMAX = 64
)

var sigChannel = make(chan os.Signal, 1)
var signalHandlers = make(map[os.Signal][]SignalHandler)

func InitSignals(signals []os.Signal) {
	signal.Notify(sigChannel, signals...)

	for _, sig := range signals {
		signalHandlers[sig] = make([]SignalHandler, 0)
	}

	go func() {
		for {
			sig := <-sigChannel

			for _, handler := range signalHandlers[sig] {
				handler(sig)
			}
		}
	}()
}

func RealTimeSignal(num int) os.Signal {
	if num < 0 || num > SIGRTMAX-SIGRTMIN {
		return nil
	}
	return syscall.Signal(SIGRTMIN + num)
}

func AddSignalHandler(signals []os.Signal, handler SignalHandler) {
	for _, sig := range signals {
		signalHandlers[sig] = append(signalHandlers[sig], handler)
	}
}
