package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAllowedOrigins(t *testing.T) {
	if err := validateAllowedOrigins([]string{"http://localhost:3000", "https://client.example"}); err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{"*", "null", "https://*.example", "ftp://client.example", "https://client.example/", "https://client.example/path", "https://client.example?x=1", "https://user@client.example"} {
		t.Run(origin, func(t *testing.T) {
			if err := validateAllowedOrigins([]string{origin}); err == nil || !strings.Contains(err.Error(), "exact http or https origin") {
				t.Fatalf("origin %q accepted or returned unstable error: %v", origin, err)
			}
		})
	}
}

func TestAcquireAPILockRejectsSecondOwner(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "api")
	first, err := acquireAPILock(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := acquireAPILock(directory)
	if second != nil {
		second.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "another sudoku api process") {
		t.Fatalf("second owner was not rejected clearly: %v", err)
	}
}
