package logr

import "testing"

func TestInfo(t *testing.T) {
	Info(12, "rao", []string{"q", "w", "e"})
}

func TestSetup(t *testing.T) {
	Setup(Config{FilePath: "runtime/consoles", FileNamePrefix: "console"})
	Info(12, "rao", []string{"q", "w", "e"})
}
