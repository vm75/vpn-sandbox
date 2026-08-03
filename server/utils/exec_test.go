package utils

import (
	"testing"
	"time"
)

func TestRunCommandLoggedDoesNotWaitForBackgroundChildren(t *testing.T) {
	InitLog(t.TempDir() + "/command.log")
	t.Cleanup(func() {
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	})

	started := time.Now()
	if err := RunCommandLogged(false, "/bin/sh", "-c", "sleep 1 &"); err != nil {
		t.Fatalf("RunCommandLogged() error = %v", err)
	}

	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("RunCommandLogged() waited %v for a background child", elapsed)
	}
}
