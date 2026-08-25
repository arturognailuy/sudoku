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
	for i := range manifest.Puzzles {
		puzzle := &manifest.Puzzles[i]
		puzzle.PuzzleHash = hash([]byte(puzzle.Puzzle))
		if puzzle.SourceCategory == "" {
			puzzle.SourceCategory = "external"
		}
		if puzzle.SourceID == "" {
			puzzle.SourceID = "repository-test-fixture"
		}
		if puzzle.License == "" {
			puzzle.License = "project-license"
		}
		if puzzle.Redistribution == "" {
			puzzle.Redistribution = "permitted"
		}
		if puzzle.CollectionMethod == "" {
			puzzle.CollectionMethod = "test fixture"
		}
		if puzzle.Split == "" {
			puzzle.Split = "exploratory"
		}
	}
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
	writeManifest(t, manifestPath, Manifest{Version: Version, Name: "pilot", Puzzles: []Puzzle{{ID: "known-1", Puzzle: knownPuzzle}}})

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
	if err := os.WriteFile(path, []byte(`{"version":2,"name":"bad","extra":true,"puzzles":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field error = %v", err)
	}
}

func TestPrepareManifestNormalizesAndHashes(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "candidate.json")
	output := filepath.Join(directory, "manifest.json")
	candidate := Manifest{Version: Version, Name: "candidate", Puzzles: []Puzzle{{
		ID: "known-1", Puzzle: strings.ReplaceAll(knownPuzzle, ".", "0"),
		SourceCategory: "external", SourceID: "fixture-1", License: "CC0-1.0",
		Redistribution: "permitted", CollectionMethod: "checked-in test fixture", Split: "held-out",
	}}}
	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, err := PrepareManifest(input, output)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Puzzles[0].Puzzle != knownPuzzle || manifest.Puzzles[0].PuzzleHash != hash([]byte(knownPuzzle)) {
		t.Fatalf("manifest was not normalized and hashed: %+v", manifest.Puzzles[0])
	}
	if _, _, err := loadManifest(output); err != nil {
		t.Fatalf("prepared manifest is not runnable: %v", err)
	}
}

func TestPrepareManifestRejectsDuplicateNormalizedContent(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "candidate.json")
	candidate := Manifest{Version: Version, Name: "duplicates", Puzzles: []Puzzle{
		{ID: "dots", Puzzle: knownPuzzle, SourceCategory: "external", SourceID: "dots", License: "CC0", Redistribution: "permitted", CollectionMethod: "fixture", Split: "exploratory"},
		{ID: "zeros", Puzzle: strings.ReplaceAll(knownPuzzle, ".", "0"), SourceCategory: "external", SourceID: "zeros", License: "CC0", Redistribution: "permitted", CollectionMethod: "fixture", Split: "held-out"},
	}}
	data, _ := json.Marshal(candidate)
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareManifest(input, filepath.Join(directory, "manifest.json")); err == nil || !strings.Contains(err.Error(), "duplicate puzzle hash") {
		t.Fatalf("error = %v, want duplicate puzzle hash", err)
	}
}

func TestLoadManifestRequiresProvenance(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "manifest.json")
	manifest := Manifest{Version: Version, Name: "missing", Puzzles: []Puzzle{{ID: "known", Puzzle: knownPuzzle}}}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "puzzle_hash") {
		t.Fatalf("error = %v, want strict corpus metadata rejection", err)
	}
}

func TestLoadManifestValidatesGeneratedMetadata(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "manifest.json")
	manifest := Manifest{Version: Version, Name: "generated", Puzzles: []Puzzle{{
		ID: "generated-1", Puzzle: knownPuzzle, SourceCategory: "generated",
	}}}
	writeManifest(t, path, manifest)
	if _, _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "requires generator metadata") {
		t.Fatalf("error = %v, want generated metadata rejection", err)
	}

	manifest.Puzzles[0].Generator = &GeneratorMetadata{
		RequestedDifficulty: "hard", Configuration: "rounds=10 timeout=1s",
		Attempt: 2, ElapsedMilliseconds: 125, ClueCount: 30,
		ClassificationOutcome: string(solver.ClassificationSolved),
	}
	writeManifest(t, path, manifest)
	if _, _, err := loadManifest(path); err != nil {
		t.Fatalf("valid generated metadata rejected: %v", err)
	}
}
