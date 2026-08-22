package cmd

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
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

func TestAcquireAPILockCanBeReacquiredAfterClose(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "api")
	first, err := acquireAPILock(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := acquireAPILock(directory)
	if err != nil {
		t.Fatalf("reacquire lock: %v", err)
	}
	second.Close()
}

func TestRunAPIRejectsUnsafeConfigurationBeforeStartup(t *testing.T) {
	command := &cobra.Command{}
	command.SetContext(context.Background())
	base := apiConfig{
		listen:          "127.0.0.1:8080",
		readTimeout:     time.Second,
		writeTimeout:    time.Second,
		idleTimeout:     time.Second,
		shutdownTimeout: time.Second,
	}
	for _, test := range []struct {
		name   string
		mutate func(*apiConfig)
		want   string
	}{
		{name: "invalid listen", mutate: func(c *apiConfig) { c.listen = "not-an-address" }, want: "invalid --listen address"},
		{name: "remote without token", mutate: func(c *apiConfig) { c.listen = "0.0.0.0:8080" }, want: "--auth-token is required"},
		{name: "token newline", mutate: func(c *apiConfig) { c.token = "bad\ntoken" }, want: "--auth-token contains invalid characters"},
		{name: "invalid origin", mutate: func(c *apiConfig) { c.origins = []string{"https://client.example/path"} }, want: "invalid --allowed-origin"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := base
			test.mutate(&config)
			err := runAPI(command, config)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}
