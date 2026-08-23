package calibration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnailuy/sudoku/solver"
)

const knownPuzzle = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

func writeManifest(t *testing.T, path string, manifest Manifest) {
	t.Helper()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunAppendsOnceAndResumes(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "manifest.json")
	output := filepath.Join(directory, "run")
	writeManifest(t, manifestPath, Manifest{Version: Version, Name: "pilot", Puzzles: []Puzzle{{ID: "known-1", Puzzle: knownPuzzle, Source: "fixture"}}})

	first, err := Run(manifestPath, output, solver.NewStore())
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if first.Appended != 1 || !first.Report.Complete || first.Report.Observed != 1 {
		t.Fatalf("unexpected first result: %+v", first)
	}
	observationsBefore, err := os.ReadFile(filepath.Join(output, "observations.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	reportBefore, err := os.ReadFile(filepath.Join(output, "report.json"))
	if err != nil {
		t.Fatal(err)
	}

	second, err := Run(manifestPath, output, solver.NewStore())
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if second.Appended != 0 || second.Report.Observed != 1 {
		t.Fatalf("unexpected resumed result: %+v", second)
	}
	observationsAfter, _ := os.ReadFile(filepath.Join(output, "observations.jsonl"))
	reportAfter, _ := os.ReadFile(filepath.Join(output, "report.json"))
	if string(observationsAfter) != string(observationsBefore) {
		t.Fatal("resume changed append-only observations")
	}
	if string(reportAfter) != string(reportBefore) {
		t.Fatal("derived report is not deterministic")
	}
}

func TestRunRecoversLaggingCheckpoint(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "manifest.json")
	output := filepath.Join(directory, "run")
	writeManifest(t, manifestPath, Manifest{Version: Version, Name: "pilot", Puzzles: []Puzzle{{ID: "known-1", Puzzle: knownPuzzle}}})
	result, err := Run(manifestPath, output, solver.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(filepath.Join(output, "checkpoint.json"), checkpoint{Version: Version, ManifestHash: result.ManifestHash, NextIndex: 0}); err != nil {
		t.Fatal(err)
	}
	resumed, err := Run(manifestPath, output, solver.NewStore())
	if err != nil {
		t.Fatalf("resume lagging checkpoint: %v", err)
	}
	if resumed.Appended != 0 {
		t.Fatalf("appended = %d, want 0", resumed.Appended)
	}
	data, _ := os.ReadFile(filepath.Join(output, "checkpoint.json"))
	if !strings.Contains(string(data), `"next_index": 1`) {
		t.Fatalf("checkpoint was not rebuilt: %s", data)
	}
}

func TestRunRejectsChangedManifest(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "manifest.json")
	output := filepath.Join(directory, "run")
	manifest := Manifest{Version: Version, Name: "pilot", Puzzles: []Puzzle{{ID: "known-1", Puzzle: knownPuzzle}}}
	writeManifest(t, manifestPath, manifest)
	if _, err := Run(manifestPath, output, solver.NewStore()); err != nil {
		t.Fatal(err)
	}
	manifest.Name = "changed"
	writeManifest(t, manifestPath, manifest)
	if _, err := Run(manifestPath, output, solver.NewStore()); err == nil || !strings.Contains(err.Error(), "immutable manifest") {
		t.Fatalf("error = %v, want immutable manifest rejection", err)
	}
}

func TestLoadManifestRejectsDuplicateIDsAndUnknownFields(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "manifest.json")
	writeManifest(t, path, Manifest{Version: Version, Name: "bad", Puzzles: []Puzzle{{ID: "same", Puzzle: knownPuzzle}, {ID: "same", Puzzle: knownPuzzle}}})
	if _, _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("duplicate error = %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"name":"bad","extra":true,"puzzles":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field error = %v", err)
	}
}
