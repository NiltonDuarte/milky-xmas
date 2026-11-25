package logfile

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupAppendsToTheFileAndKeepsAnOldOneAside(t *testing.T) {
	defer log.SetOutput(os.Stderr)
	defer log.SetFlags(log.LstdFlags)
	dir := t.TempDir()
	path := filepath.Join(dir, "logs", "bot.log")

	closeLog, err := Setup(path)
	if err != nil {
		t.Fatal(err)
	}
	log.Print("first line")
	closeLog()
	closeLog, err = Setup(path)
	if err != nil {
		t.Fatal(err)
	}
	log.Print("second line")
	closeLog()
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "first line") || !strings.Contains(string(data), "second line") {
		t.Fatalf("log = %q, %v, want both lines appended", data, err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Errorf("log permissions = %v, want owner only", info.Mode().Perm())
	}

	if err := os.WriteFile(path, make([]byte, maxSize+1), 0o600); err != nil {
		t.Fatal(err)
	}
	closeLog, err = Setup(path)
	if err != nil {
		t.Fatal(err)
	}
	log.Print("after rotation")
	closeLog()
	if info, err := os.Stat(path + ".1"); err != nil || info.Size() != maxSize+1 {
		t.Errorf("rotated log = %v, %v, want the old file", info, err)
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "after rotation") || len(data) > 200 {
		t.Errorf("new log = %q, want only the new line", data)
	}
}

func TestSetupWithoutAPathChangesNothing(t *testing.T) {
	closeLog, err := Setup("")
	if err != nil || closeLog() != nil {
		t.Fatalf("Setup(\"\") = %v", err)
	}
	if log.Writer() != io.Writer(os.Stderr) {
		t.Error("logger output changed")
	}
}
