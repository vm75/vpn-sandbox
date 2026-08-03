package utils

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type SignalHandler func(os.Signal)

const (
	SIGRTMIN = 34
	SIGRTMAX = 64
)

// Leave room for an up/down transition while the previous lifecycle action is
// still completing. os/signal uses a non-blocking send and drops a signal when
// this channel is full.
var sigChannel = make(chan os.Signal, 16)
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

func SignalProcess(pid int, signal os.Signal) error {
	sigStr := signal.String()
	if sigStr == "" {
		return errors.New("invalid signal")
	}
	_, err := RunCommand(UseSudo, "/bin/kill", "-"+sigStr[7:], strconv.Itoa(pid))
	return err
}
