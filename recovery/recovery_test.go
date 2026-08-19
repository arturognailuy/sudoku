package recovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteDiscoverAndDeletePrivateRecords(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "state", "sudoku", "recovery")
	store := NewStore(directory)
	first, _ := NewID()
	second, _ := NewID()
	if err := store.Write(first, "first", []byte("session-one")); err != nil {
		t.Fatal(err)
	}
	store.now = func() time.Time { return time.Now().Add(time.Minute) }
	if err := store.Write(second, "second", []byte("session-two")); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(directory)
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("directory mode=%v err=%v", info.Mode().Perm(), err)
	}
	for _, id := range []string{first, second} {
		info, err = os.Stat(filepath.Join(directory, id+".json"))
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("record mode=%v err=%v", info.Mode().Perm(), err)
		}
	}
	records, err := store.Discover(func(data []byte) error {
		if !strings.HasPrefix(string(data), "session-") {
			t.Fatal("validator received wrong data")
		}
		return nil
	})
	if err != nil || len(records) != 2 || records[0].ID != second {
		t.Fatalf("records=%+v err=%v", records, err)
	}
	if err := store.Delete(second); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, second+".json")); !os.IsNotExist(err) {
		t.Fatalf("record was not deleted: %v", err)
	}
}

func TestDiscoverPrunesMalformedExpiredAndRejectedRecords(t *testing.T) {
	directory := t.TempDir()
	store := NewStore(directory)
	store.now = func() time.Time { return time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC) }
	expired, _ := NewID()
	if err := store.Write(expired, "old", []byte("old")); err != nil {
		t.Fatal(err)
	}
	store.now = func() time.Time { return time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC) }
	rejected, _ := NewID()
	if err := store.Write(rejected, "bad", []byte("reject")); err != nil {
		t.Fatal(err)
	}
	malformed := filepath.Join(directory, strings.Repeat("a", 32)+".json")
	if err := os.WriteFile(malformed, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}

	records, err := store.Discover(func(data []byte) error {
		if string(data) == "reject" {
			return os.ErrInvalid
		}
		return nil
	})
	if err != nil || len(records) != 0 {
		t.Fatalf("records=%+v err=%v", records, err)
	}
	for _, path := range []string{filepath.Join(directory, expired+".json"), filepath.Join(directory, rejected+".json"), malformed} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("invalid record remains: %s", path)
		}
	}
}

func TestDiscoveryIgnoresSymlinksAndWriteRejectsSymlinkDirectory(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	store := NewStore(target)
	id, _ := NewID()
	outside := filepath.Join(root, "outside")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, id+".json")); err != nil {
		t.Fatal(err)
	}
	records, err := store.Discover(nil)
	if err != nil || len(records) != 0 {
		t.Fatalf("records=%v err=%v", records, err)
	}

	link := filepath.Join(root, "linked-directory")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := NewStore(link).Write(id, "", []byte("session")); err == nil {
		t.Fatal("write followed a symlink directory")
	}
}

func TestDefaultDirectoryUsesXDGStateHomeAndFallback(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/sudoku-state-test")
	path, err := DefaultDirectory()
	if err != nil || path != "/tmp/sudoku-state-test/sudoku/recovery" {
		t.Fatalf("path=%q err=%v", path, err)
	}
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "/tmp/sudoku-home-test")
	path, err = DefaultDirectory()
	if err != nil || path != "/tmp/sudoku-home-test/.local/state/sudoku/recovery" {
		t.Fatalf("path=%q err=%v", path, err)
	}
}

func TestDiscoverPrunesUnsupportedVersionAndUnsafeLabel(t *testing.T) {
	directory := t.TempDir()
	store := NewStore(directory)
	for index, mutate := range []func(map[string]any){
		func(value map[string]any) { value["version"] = float64(99) },
		func(value map[string]any) { value["source"] = "unsafe\nlabel" },
	} {
		id, _ := NewID()
		if err := store.Write(id, "safe", []byte("session")); err != nil {
			t.Fatal(err)
		}
		path := store.path(id)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var value map[string]any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		mutate(value)
		data, _ = json.Marshal(value)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		records, err := store.Discover(nil)
		if err != nil || len(records) != 0 {
			t.Fatalf("case %d records=%v err=%v", index, records, err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("case %d invalid record remains", index)
		}
	}
}
