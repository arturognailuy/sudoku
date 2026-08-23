// Package calibration provides deterministic, resumable difficulty measurements.
package calibration

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

const Version = 1

// Manifest is an immutable, ordered corpus definition.
type Manifest struct {
	Version int      `json:"version"`
	Name    string   `json:"name"`
	Puzzles []Puzzle `json:"puzzles"`
}

// Puzzle is one stable corpus member.
type Puzzle struct {
	ID                 string `json:"id"`
	Puzzle             string `json:"puzzle"`
	Source             string `json:"source,omitempty"`
	ExpectedDifficulty string `json:"expected_difficulty,omitempty"`
}

// Observation is one append-only classification result.
type Observation struct {
	Version            int                          `json:"version"`
	Index              int                          `json:"index"`
	ManifestHash       string                       `json:"manifest_hash"`
	ID                 string                       `json:"id"`
	PuzzleHash         string                       `json:"puzzle_hash"`
	Source             string                       `json:"source,omitempty"`
	ExpectedDifficulty string                       `json:"expected_difficulty,omitempty"`
	Outcome            solver.ClassificationOutcome `json:"outcome"`
	Difficulty         string                       `json:"difficulty"`
	Score              int                          `json:"score"`
	MaxTechnique       string                       `json:"max_technique"`
	MoveCount          int                          `json:"move_count"`
	TraceHash          string                       `json:"trace_hash"`
	RepeatCount        int                          `json:"repeat_count"`
	MatchesExpected    *bool                        `json:"matches_expected,omitempty"`
}

type checkpoint struct {
	Version      int    `json:"version"`
	ManifestHash string `json:"manifest_hash"`
	NextIndex    int    `json:"next_index"`
}

// Report is deterministically derived from an observation log.
type Report struct {
	Version      int            `json:"version"`
	ManifestName string         `json:"manifest_name"`
	ManifestHash string         `json:"manifest_hash"`
	Observed     int            `json:"observed"`
	Total        int            `json:"total"`
	Complete     bool           `json:"complete"`
	ByOutcome    map[string]int `json:"by_outcome"`
	ByDifficulty map[string]int `json:"by_difficulty"`
	Expected     int            `json:"expected"`
	Matched      int            `json:"matched"`
}

// Result describes the durable files produced by Run.
type Result struct {
	ManifestHash string
	Appended     int
	Report       Report
}

// Run resumes or starts a measurement run in outputDir.
func Run(manifestPath, outputDir string, store solver.Store) (Result, error) {
	manifest, manifestHash, err := loadManifest(manifestPath)
	if err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory: %w", err)
	}

	observationsPath := filepath.Join(outputDir, "observations.jsonl")
	observations, err := readObservations(observationsPath, manifest, manifestHash)
	if err != nil {
		return Result{}, err
	}
	checkpointPath := filepath.Join(outputDir, "checkpoint.json")
	if err := validateCheckpoint(checkpointPath, manifestHash, len(observations)); err != nil {
		return Result{}, err
	}
	// The observation log is authoritative. Rebuild a missing or lagging
	// checkpoint before resuming after an interruption.
	if err := writeJSONAtomic(checkpointPath, checkpoint{Version: Version, ManifestHash: manifestHash, NextIndex: len(observations)}); err != nil {
		return Result{}, err
	}

	log, err := os.OpenFile(observationsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return Result{}, fmt.Errorf("open observation log: %w", err)
	}
	defer log.Close()

	appended := 0
	for index := len(observations); index < len(manifest.Puzzles); index++ {
		observation, err := measure(index, manifestHash, manifest.Puzzles[index], store)
		if err != nil {
			return Result{}, err
		}
		line, err := json.Marshal(observation)
		if err != nil {
			return Result{}, fmt.Errorf("encode observation %d: %w", index, err)
		}
		if _, err := log.Write(append(line, '\n')); err != nil {
			return Result{}, fmt.Errorf("append observation %d: %w", index, err)
		}
		if err := log.Sync(); err != nil {
			return Result{}, fmt.Errorf("sync observation %d: %w", index, err)
		}
		observations = append(observations, observation)
		appended++
		if err := writeJSONAtomic(checkpointPath, checkpoint{Version: Version, ManifestHash: manifestHash, NextIndex: index + 1}); err != nil {
			return Result{}, err
		}
	}

	report := deriveReport(manifest, manifestHash, observations)
	if err := writeJSONAtomic(filepath.Join(outputDir, "report.json"), report); err != nil {
		return Result{}, err
	}
	if err := writeAtomic(filepath.Join(outputDir, "report.md"), []byte(markdownReport(report)), 0o644); err != nil {
		return Result{}, err
	}
	return Result{ManifestHash: manifestHash, Appended: appended, Report: report}, nil
}

func loadManifest(path string) (Manifest, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, "", fmt.Errorf("read manifest: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, "", fmt.Errorf("decode manifest: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return Manifest{}, "", err
	}
	if manifest.Version != Version {
		return Manifest{}, "", fmt.Errorf("unsupported manifest version %d", manifest.Version)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return Manifest{}, "", errors.New("manifest name is required")
	}
	seen := make(map[string]struct{}, len(manifest.Puzzles))
	seenPuzzles := make(map[string]struct{}, len(manifest.Puzzles))
	validDifficulties := map[string]bool{"": true, "easy": true, "medium": true, "hard": true, "expert": true, "evil": true}
	for i, puzzle := range manifest.Puzzles {
		if strings.TrimSpace(puzzle.ID) == "" {
			return Manifest{}, "", fmt.Errorf("puzzle %d: id is required", i)
		}
		if _, duplicate := seen[puzzle.ID]; duplicate {
			return Manifest{}, "", fmt.Errorf("puzzle %d: duplicate id %q", i, puzzle.ID)
		}
		seen[puzzle.ID] = struct{}{}
		if !core.IsValidSudokuString(puzzle.Puzzle) {
			return Manifest{}, "", fmt.Errorf("puzzle %d (%s): invalid Sudoku string", i, puzzle.ID)
		}
		board := core.NewEmptyBoard()
		board.FromString(puzzle.Puzzle)
		if board.ToString() != puzzle.Puzzle {
			return Manifest{}, "", fmt.Errorf("puzzle %d (%s): puzzle must use canonical dot notation", i, puzzle.ID)
		}
		if _, duplicate := seenPuzzles[puzzle.Puzzle]; duplicate {
			return Manifest{}, "", fmt.Errorf("puzzle %d (%s): duplicate puzzle", i, puzzle.ID)
		}
		seenPuzzles[puzzle.Puzzle] = struct{}{}
		if !validDifficulties[puzzle.ExpectedDifficulty] {
			return Manifest{}, "", fmt.Errorf("puzzle %d (%s): invalid expected difficulty %q", i, puzzle.ID, puzzle.ExpectedDifficulty)
		}
	}
	return manifest, hash(data), nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("decode manifest: trailing JSON value")
		}
		return fmt.Errorf("decode manifest: %w", err)
	}
	return nil
}

func readObservations(path string, manifest Manifest, manifestHash string) ([]Observation, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open observation log: %w", err)
	}
	defer file.Close()

	var observations []Observation
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		index := len(observations)
		var observation Observation
		if err := json.Unmarshal(scanner.Bytes(), &observation); err != nil {
			return nil, fmt.Errorf("decode observation %d: %w", index, err)
		}
		if index >= len(manifest.Puzzles) {
			return nil, fmt.Errorf("observation %d exceeds manifest length", index)
		}
		puzzle := manifest.Puzzles[index]
		if observation.Version != Version || observation.Index != index || observation.ManifestHash != manifestHash || observation.ID != puzzle.ID || observation.PuzzleHash != hash([]byte(puzzle.Puzzle)) {
			return nil, fmt.Errorf("observation %d does not match immutable manifest prefix", index)
		}
		observations = append(observations, observation)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read observation log: %w", err)
	}
	return observations, nil
}

func validateCheckpoint(path, manifestHash string, observed int) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read checkpoint: %w", err)
	}
	var value checkpoint
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("decode checkpoint: %w", err)
	}
	if value.Version != Version || value.ManifestHash != manifestHash || value.NextIndex < 0 || value.NextIndex > observed {
		return errors.New("checkpoint does not match the immutable manifest and observation log")
	}
	return nil
}

func measure(index int, manifestHash string, puzzle Puzzle, store solver.Store) (Observation, error) {
	board := core.NewEmptyBoard()
	board.FromString(puzzle.Puzzle)
	classification := solver.ClassifyPuzzle(store, board)
	repeated := solver.ClassifyPuzzle(store, board)
	if classificationSignature(classification) != classificationSignature(repeated) {
		return Observation{}, fmt.Errorf("puzzle %d (%s): repeated classification was not identical", index, puzzle.ID)
	}
	observation := Observation{
		Version: Version, Index: index, ManifestHash: manifestHash, ID: puzzle.ID,
		PuzzleHash: hash([]byte(puzzle.Puzzle)), Source: puzzle.Source,
		ExpectedDifficulty: puzzle.ExpectedDifficulty, Outcome: classification.Outcome,
		Difficulty: classification.Difficulty, Score: classification.Score,
		MaxTechnique: classification.MaxTechnique, MoveCount: len(classification.Moves),
		TraceHash: traceHash(classification.Moves), RepeatCount: 2,
	}
	if puzzle.ExpectedDifficulty != "" {
		matched := classification.Difficulty == puzzle.ExpectedDifficulty
		observation.MatchesExpected = &matched
	}
	return observation, nil
}

func classificationSignature(classification solver.Classification) string {
	return fmt.Sprintf("%s\x00%s\x00%d\x00%s\x00%d\x00%s", classification.Outcome, classification.Difficulty, classification.Score, classification.MaxTechnique, len(classification.Moves), traceHash(classification.Moves))
}

func traceHash(moves []solver.Move) string {
	type traceMove struct {
		Row             int    `json:"row"`
		Column          int    `json:"column"`
		Value           int    `json:"value"`
		Technique       string `json:"technique"`
		Reason          string `json:"reason"`
		EliminationOnly bool   `json:"elimination_only"`
	}
	trace := make([]traceMove, len(moves))
	for i, move := range moves {
		trace[i] = traceMove{move.Cell.Position.Row, move.Cell.Position.Column, move.Cell.Value, move.Technique, move.Reason, move.EliminationOnly}
	}
	data, _ := json.Marshal(trace)
	return hash(data)
}

func deriveReport(manifest Manifest, manifestHash string, observations []Observation) Report {
	report := Report{Version: Version, ManifestName: manifest.Name, ManifestHash: manifestHash, Observed: len(observations), Total: len(manifest.Puzzles), Complete: len(observations) == len(manifest.Puzzles), ByOutcome: map[string]int{}, ByDifficulty: map[string]int{}}
	for _, observation := range observations {
		report.ByOutcome[string(observation.Outcome)]++
		report.ByDifficulty[observation.Difficulty]++
		if observation.MatchesExpected != nil {
			report.Expected++
			if *observation.MatchesExpected {
				report.Matched++
			}
		}
	}
	return report
}

func markdownReport(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Difficulty Measurement Report\n\n- Manifest: `%s`\n- Hash: `%s`\n- Progress: %d/%d\n- Complete: %t\n", report.ManifestName, report.ManifestHash, report.Observed, report.Total, report.Complete)
	if report.Expected > 0 {
		fmt.Fprintf(&builder, "- Expected matches: %d/%d\n", report.Matched, report.Expected)
	}
	for _, section := range []struct {
		name   string
		values map[string]int
	}{{"Outcomes", report.ByOutcome}, {"Difficulties", report.ByDifficulty}} {
		fmt.Fprintf(&builder, "\n## %s\n\n", section.name)
		keys := make([]string, 0, len(section.values))
		for key := range section.values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&builder, "- %s: %d\n", key, section.values[key])
		}
	}
	return builder.String()
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", filepath.Base(path), err)
	}
	return writeAtomic(path, append(data, '\n'), 0o644)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temporary %s: %w", filepath.Base(path), err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", filepath.Base(path), err)
	}
	return nil
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
