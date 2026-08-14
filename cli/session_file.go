package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MaxSessionFileSize bounds restore input before the game engine parses it.
const MaxSessionFileSize int64 = 1 << 20

// ErrSessionTooLarge identifies a restore file that exceeds MaxSessionFileSize.
var ErrSessionTooLarge = errors.New("session file is too large")

// ReadSessionFile reads a bounded serialized session from path.
func ReadSessionFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session %q: %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, MaxSessionFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("read session %q: %w", path, err)
	}
	if int64(len(data)) > MaxSessionFileSize {
		return nil, fmt.Errorf("read session %q: %w (maximum %d bytes)", path, ErrSessionTooLarge, MaxSessionFileSize)
	}
	return data, nil
}

// WriteSessionFile atomically replaces path with a user-readable serialized
// session. The temporary file is kept beside the destination so rename stays
// on the same filesystem.
func WriteSessionFile(path string, data []byte) (err error) {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".sudoku-session-*")
	if err != nil {
		return fmt.Errorf("create temporary session beside %q: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
	}()

	if err = temporary.Chmod(0600); err != nil {
		return fmt.Errorf("protect temporary session for %q: %w", path, err)
	}
	if _, err = temporary.Write(data); err != nil {
		return fmt.Errorf("write temporary session for %q: %w", path, err)
	}
	if err = temporary.Sync(); err != nil {
		return fmt.Errorf("flush temporary session for %q: %w", path, err)
	}
	if err = temporary.Close(); err != nil {
		return fmt.Errorf("close temporary session for %q: %w", path, err)
	}
	if err = os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace session %q: %w", path, err)
	}

	return nil
}
