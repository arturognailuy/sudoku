package game

import (
	"fmt"
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/message"
	"github.com/gnailuy/sudoku/solver"
)

func intersectionSteps(move *solver.Move) []HintStep {
	evidence := move.Evidence
	conclusion := HintConclusion{Placement: placement(move), Eliminations: evidence.Eliminations}
	if len(evidence.Units) != 2 || len(evidence.Premises) < 2 {
		return evidenceSteps(move, conclusion)
	}
	digit := evidence.Premises[0].Values[0]
	source, intersection := evidence.Units[0], evidence.Units[1]
	if move.Technique == "pointing-pair" && source.Kind == solver.UnitBox && len(evidence.Premises) == 2 &&
		conclusion.Placement != nil {
		first := solver.CandidateRef{Position: evidence.Premises[0].Position, Value: digit}
		second := solver.CandidateRef{Position: evidence.Premises[1].Position, Value: digit}
		removed := solver.CandidateRef{Position: conclusion.Placement.Position, Value: digit}
		for _, ref := range append(append([]solver.CandidateRef(nil), evidence.RuledOut...), evidence.Eliminations...) {
			if ref.Position == conclusion.Placement.Position && ref.Value == digit {
				removed = ref
				break
			}
		}
		boxTarget, _ := unitTarget(&source)
		columnTarget, _ := unitTarget(&intersection)
		firstMarks := []HintMark{
			{Target: boxTarget, Role: HintRoleFocus},
			{Target: candidateTarget(first), Role: HintRolePremise},
		}
		pairMarks := append(append([]HintMark(nil), firstMarks...),
			HintMark{Target: candidateTarget(second), Role: HintRolePremise},
			HintMark{Target: columnTarget, Role: HintRoleFocus},
		)
		consequenceMarks := append(
			append([]HintMark(nil), pairMarks...),
			HintMark{Target: candidateTarget(removed), Role: HintRoleEliminated},
		)
		applyMarks := append(
			append([]HintMark(nil), consequenceMarks...),
			HintMark{
				Target: candidateTarget(
					solver.CandidateRef{Position: conclusion.Placement.Position, Value: conclusion.Placement.Value},
				),
				Role: HintRoleConclusion,
			},
		)
		firstMessage := hintDescriptor(
			"hint.pointing-pair.first",
			unitsArgument("source_unit", []solver.UnitRef{source}),
			digitArgument("digit", digit),
			cellArgument("cell", first.Position),
		)
		secondMessage := hintDescriptor(
			"hint.pointing-pair.second",
			digitArgument("digit", digit),
			unitsArgument("source_unit", []solver.UnitRef{source}),
			cellArgument("cell", second.Position),
			unitsArgument("intersection_unit", []solver.UnitRef{intersection}),
		)
		consequenceMessage := hintDescriptor(
			"hint.pointing-pair.consequence",
			digitArgument("digit", digit),
			unitsArgument("source_unit", []solver.UnitRef{source}),
			unitsArgument("intersection_unit", []solver.UnitRef{intersection}),
			message.Argument{
				Name: "candidate",
				Kind: message.ArgumentCandidate,
				Candidate: &message.Candidate{
					Cell:  message.Cell{Row: removed.Position.Row + 1, Column: removed.Position.Column + 1},
					Digit: removed.Value,
				},
			},
		)
		applyMessage := hintDescriptor(
			"hint.pointing-pair.apply",
			cellArgument("cell", conclusion.Placement.Position),
			digitArgument("digit", conclusion.Placement.Value),
			strategyArgument("strategy", move.Technique),
		)
		return []HintStep{
			{
				ID:        "inspect-" + compactCellName(first.Position) + "-candidate",
				Kind:      HintStepObserve,
				Message:   firstMessage,
				Narration: []HintNarrationChunk{narrationMessage(firstMessage)},
				Marks:     firstMarks,
			},
			{
				ID:        "complete-pointing-pair",
				Kind:      HintStepExplain,
				Message:   secondMessage,
				Narration: []HintNarrationChunk{narrationMessage(secondMessage)},
				Marks:     pairMarks,
			},
			{
				ID:        "remove-" + compactCellName(removed.Position) + fmt.Sprintf("-%d", removed.Value),
				Kind:      HintStepExplain,
				Message:   consequenceMessage,
				Narration: []HintNarrationChunk{narrationMessage(consequenceMessage)},
				Marks:     consequenceMarks,
			},
			{
				ID: "place-" + compactCellName(
					conclusion.Placement.Position,
				) + fmt.Sprintf(
					"-%d",
					conclusion.Placement.Value,
				),
				Kind:      HintStepApply,
				Message:   applyMessage,
				Narration: actionNarration(applyMessage),
				Marks:     applyMarks,
				Effect:    conclusionEffect(conclusion),
			},
		}
	}
	premiseRefs := make([]solver.CandidateRef, 0, len(evidence.Premises))
	premiseCells := make([]core.Position, 0, len(evidence.Premises))
	for _, premise := range evidence.Premises {
		premiseRefs = append(premiseRefs, solver.CandidateRef{Position: premise.Position, Value: digit})
		premiseCells = append(premiseCells, premise.Position)
	}
	observe := hintDescriptor(
		"hint.intersection-confinement.observe",
		unitsArgument("source_unit", []solver.UnitRef{source}),
		digitArgument("digit", digit),
		unitsArgument("intersection_unit", []solver.UnitRef{intersection}),
		cellsArgument("cells", premiseCells),
	)
	explain := hintDescriptor(
		"hint.intersection-confinement.explain",
		digitArgument("digit", digit),
		unitsArgument("source_unit", []solver.UnitRef{source}),
		unitsArgument("intersection_unit", []solver.UnitRef{intersection}),
	)
	var apply message.Descriptor
	if conclusion.Placement != nil {
		apply = hintDescriptor(
			"hint.intersection-confinement.place",
			digitArgument("digit", conclusion.Placement.Value),
			cellArgument("cell", conclusion.Placement.Position),
			strategyArgument("strategy", move.Technique),
		)
	} else {
		positions := make([]core.Position, len(evidence.Eliminations))
		for i, ref := range evidence.Eliminations {
			positions[i] = ref.Position
		}
		apply = hintDescriptor("hint.intersection-confinement.apply", digitArgument("digit", digit),
			cellsArgument("cells", positions), strategyArgument("strategy", move.Technique))
	}
	premiseMarks := []HintMark{}
	if target, ok := unitTarget(&source); ok {
		premiseMarks = append(premiseMarks, HintMark{Target: target, Role: HintRoleFocus})
	}
	for _, ref := range premiseRefs {
		premiseMarks = append(premiseMarks, HintMark{Target: candidateTarget(ref), Role: HintRolePremise})
	}
	explainMarks := append([]HintMark(nil), premiseMarks...)
	if target, ok := unitTarget(&intersection); ok {
		explainMarks = append(explainMarks, HintMark{Target: target, Role: HintRoleFocus})
	}
	return []HintStep{
		{
			ID:        "find-intersection",
			Kind:      HintStepObserve,
			Message:   observe,
			Narration: []HintNarrationChunk{narrationMessage(observe)},
			Marks:     premiseMarks,
		},
		{
			ID:        "explain-intersection",
			Kind:      HintStepExplain,
			Message:   explain,
			Narration: []HintNarrationChunk{narrationMessage(explain)},
			Marks:     explainMarks,
		},
		{
			ID:        "apply-intersection",
			Kind:      conclusionKind(conclusion),
			Message:   apply,
			Narration: actionNarration(apply),
			Marks:     append(explainMarks, conclusionMarks(conclusion)...),
			Effect:    conclusionEffect(conclusion),
		},
	}
}

func fishSteps(move *solver.Move, _ string) []HintStep {
	evidence := move.Evidence
	marks := unitAndPremiseMarks(evidence)
	conclusion := HintConclusion{Placement: placement(move), Eliminations: evidence.Eliminations}
	size := map[string]int{"x-wing": 2, "swordfish": 3, "jellyfish": 4}[move.Technique]
	if size != 0 && len(evidence.Units) == size*2 && len(evidence.Premises) > 0 &&
		len(evidence.Premises[0].Values) == 1 &&
		len(evidence.Eliminations) > 0 {
		baseUnits, coverUnits := evidence.Units[:size], evidence.Units[size:]
		digit := evidence.Premises[0].Values[0]
		observe := hintDescriptor(
			"hint.fish-confinement.observe",
			unitsArgument("base_units", baseUnits),
			digitArgument("digit", digit),
			unitsArgument("cover_units", coverUnits),
		)
		explain := hintDescriptor(
			"hint.fish-confinement.explain",
			unitsArgument("base_units", baseUnits),
			unitsArgument("cover_units", coverUnits),
			digitArgument("digit", digit),
		)
		positions := make([]core.Position, len(evidence.Eliminations))
		for i, ref := range evidence.Eliminations {
			positions[i] = ref.Position
		}
		if move.Technique == "x-wing" && len(evidence.Premises) == 4 && baseUnits[0].Kind == solver.UnitRow &&
			baseUnits[1].Kind == solver.UnitRow &&
			coverUnits[0].Kind == solver.UnitColumn &&
			coverUnits[1].Kind == solver.UnitColumn {
			firstBase, secondBase := baseUnits[0], baseUnits[1]
			firstMarks := make([]HintMark, 0, 2)
			patternMarks := make([]HintMark, 0, 6)
			for _, premise := range evidence.Premises[:2] {
				firstMarks = append(
					firstMarks,
					HintMark{
						Target: candidateTarget(solver.CandidateRef{Position: premise.Position, Value: digit}),
						Role:   HintRolePremise,
					},
				)
			}
			for _, premise := range evidence.Premises {
				patternMarks = append(
					patternMarks,
					HintMark{
						Target: candidateTarget(solver.CandidateRef{Position: premise.Position, Value: digit}),
						Role:   HintRolePremise,
					},
				)
			}
			for _, ref := range baseUnits {
				target, _ := unitTarget(&ref)
				patternMarks = append(patternMarks, HintMark{Target: target, Role: HintRoleFocus})
			}
			consequenceMarks := append([]HintMark(nil), patternMarks...)
			for _, ref := range coverUnits {
				target, _ := unitTarget(&ref)
				consequenceMarks = append(consequenceMarks, HintMark{Target: target, Role: HintRoleFocus})
			}
			first := hintDescriptor(
				"hint.x-wing.first",
				unitsArgument("base_unit", []solver.UnitRef{firstBase}),
				digitArgument("digit", digit),
				unitsArgument("cover_units", coverUnits),
			)
			second := hintDescriptor(
				"hint.x-wing.second",
				unitsArgument("base_unit", []solver.UnitRef{secondBase}),
				digitArgument("digit", digit),
				unitsArgument("cover_units", coverUnits),
			)
			consequence := hintDescriptor(
				"hint.x-wing.consequence",
				unitsArgument("base_units", baseUnits),
				unitsArgument("cover_units", coverUnits),
				digitArgument("digit", digit),
			)
			apply := hintDescriptor(
				"hint.x-wing.apply",
				digitArgument("digit", digit),
				cellsArgument("cells", positions),
				strategyArgument("strategy", move.Technique),
			)
			return []HintStep{
				{
					ID:        "find-first-x-wing-line",
					Kind:      HintStepObserve,
					Message:   first,
					Narration: []HintNarrationChunk{narrationMessage(first)},
					Marks:     firstMarks,
				},
				{
					ID:        "find-matching-x-wing-line",
					Kind:      HintStepExplain,
					Message:   second,
					Narration: []HintNarrationChunk{narrationMessage(second)},
					Marks:     patternMarks,
				},
				{
					ID:        "explain-x-wing-columns",
					Kind:      HintStepExplain,
					Message:   consequence,
					Narration: []HintNarrationChunk{narrationMessage(consequence)},
					Marks:     consequenceMarks,
				},
				{
					ID:        "apply-x-wing",
					Kind:      HintStepApply,
					Message:   apply,
					Narration: actionNarration(apply),
					Marks:     append(consequenceMarks, conclusionMarks(conclusion)...),
					Effect:    conclusionEffect(conclusion),
				},
			}
		}
		if move.Technique == "swordfish" && len(evidence.Premises) == 6 && baseUnits[0].Kind == solver.UnitRow &&
			baseUnits[1].Kind == solver.UnitRow &&
			baseUnits[2].Kind == solver.UnitRow &&
			coverUnits[0].Kind == solver.UnitColumn &&
			coverUnits[1].Kind == solver.UnitColumn &&
			coverUnits[2].Kind == solver.UnitColumn {
			premiseMarks := premiseCandidateMarks(evidence)
			firstMarks := append([]HintMark(nil), premiseMarks[:2]...)
			secondMarks := append([]HintMark(nil), premiseMarks[:4]...)
			for _, ref := range baseUnits[:2] {
				target, _ := unitTarget(&ref)
				secondMarks = append(secondMarks, HintMark{Target: target, Role: HintRoleFocus})
			}
			thirdMarks := append([]HintMark(nil), premiseMarks...)
			for _, ref := range baseUnits {
				target, _ := unitTarget(&ref)
				thirdMarks = append(thirdMarks, HintMark{Target: target, Role: HintRoleFocus})
			}
			consequenceMarks := append([]HintMark(nil), thirdMarks...)
			for _, ref := range coverUnits {
				target, _ := unitTarget(&ref)
				consequenceMarks = append(consequenceMarks, HintMark{Target: target, Role: HintRoleFocus})
			}
			lineCoverUnits := func(premises []solver.CandidateGroup) []solver.UnitRef {
				refs := make([]solver.UnitRef, len(premises))
				for index, premise := range premises {
					refs[index] = solver.UnitRef{Kind: solver.UnitColumn, Index: premise.Position.Column}
				}
				return refs
			}
			first := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[:1]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[:2])),
			)
			second := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[1:2]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[2:4])),
			)
			third := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[2:3]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[4:6])),
			)
			consequence := hintDescriptor(
				"hint.swordfish.consequence",
				unitsArgument("base_units", baseUnits),
				unitsArgument("cover_units", coverUnits),
				digitArgument("digit", digit),
			)
			apply := hintDescriptor(
				"hint.fish-confinement.apply",
				digitArgument("digit", digit),
				cellsArgument("cells", positions),
				strategyArgument("strategy", move.Technique),
			)
			return []HintStep{
				{
					ID:        "find-first-swordfish-row",
					Kind:      HintStepObserve,
					Message:   first,
					Narration: []HintNarrationChunk{narrationMessage(first)},
					Marks:     firstMarks,
				},
				{
					ID:        "add-second-swordfish-row",
					Kind:      HintStepExplain,
					Message:   second,
					Narration: []HintNarrationChunk{narrationMessage(second)},
					Marks:     secondMarks,
				},
				{
					ID:        "complete-swordfish-rows",
					Kind:      HintStepExplain,
					Message:   third,
					Narration: []HintNarrationChunk{narrationMessage(third)},
					Marks:     thirdMarks,
				},
				{
					ID:        "explain-swordfish-columns",
					Kind:      HintStepExplain,
					Message:   consequence,
					Narration: []HintNarrationChunk{narrationMessage(consequence)},
					Marks:     consequenceMarks,
				},
				{
					ID:        "apply-swordfish",
					Kind:      HintStepApply,
					Message:   apply,
					Narration: actionNarration(apply),
					Marks:     append(consequenceMarks, conclusionMarks(conclusion)...),
					Effect:    conclusionEffect(conclusion),
				},
			}
		}
		if move.Technique == "jellyfish" && len(evidence.Premises) == 8 && baseUnits[0].Kind == solver.UnitRow &&
			baseUnits[1].Kind == solver.UnitRow &&
			baseUnits[2].Kind == solver.UnitRow &&
			baseUnits[3].Kind == solver.UnitRow &&
			coverUnits[0].Kind == solver.UnitColumn &&
			coverUnits[1].Kind == solver.UnitColumn &&
			coverUnits[2].Kind == solver.UnitColumn &&
			coverUnits[3].Kind == solver.UnitColumn {
			premiseMarks := premiseCandidateMarks(evidence)
			lineCoverUnits := func(premises []solver.CandidateGroup) []solver.UnitRef {
				refs := make([]solver.UnitRef, len(premises))
				for index, premise := range premises {
					refs[index] = solver.UnitRef{Kind: solver.UnitColumn, Index: premise.Position.Column}
				}
				return refs
			}
			lineMarks := func(lineCount int) []HintMark {
				marks := append([]HintMark(nil), premiseMarks[:lineCount*2]...)
				if lineCount > 1 {
					for _, ref := range baseUnits[:lineCount] {
						target, _ := unitTarget(&ref)
						marks = append(marks, HintMark{Target: target, Role: HintRoleFocus})
					}
				}
				return marks
			}
			firstMarks := lineMarks(1)
			secondMarks := lineMarks(2)
			thirdMarks := lineMarks(3)
			fourthMarks := lineMarks(4)
			consequenceMarks := append([]HintMark(nil), fourthMarks...)
			for _, ref := range coverUnits {
				target, _ := unitTarget(&ref)
				consequenceMarks = append(consequenceMarks, HintMark{Target: target, Role: HintRoleFocus})
			}
			first := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[:1]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[:2])),
			)
			second := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[1:2]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[2:4])),
			)
			third := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[2:3]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[4:6])),
			)
			fourth := hintDescriptor(
				"hint.fish-confinement.observe",
				unitsArgument("base_units", baseUnits[3:4]),
				digitArgument("digit", digit),
				unitsArgument("cover_units", lineCoverUnits(evidence.Premises[6:8])),
			)
			consequence := hintDescriptor(
				"hint.jellyfish.consequence",
				unitsArgument("base_units", baseUnits),
				unitsArgument("cover_units", coverUnits),
				digitArgument("digit", digit),
			)
			apply := hintDescriptor(
				"hint.fish-confinement.apply",
				digitArgument("digit", digit),
				cellsArgument("cells", positions),
				strategyArgument("strategy", move.Technique),
			)
			return []HintStep{
				{
					ID:        "find-first-jellyfish-row",
					Kind:      HintStepObserve,
					Message:   first,
					Narration: []HintNarrationChunk{narrationMessage(first)},
					Marks:     firstMarks,
				},
				{
					ID:        "add-second-jellyfish-row",
					Kind:      HintStepExplain,
					Message:   second,
					Narration: []HintNarrationChunk{narrationMessage(second)},
					Marks:     secondMarks,
				},
				{
					ID:        "add-third-jellyfish-row",
					Kind:      HintStepExplain,
					Message:   third,
					Narration: []HintNarrationChunk{narrationMessage(third)},
					Marks:     thirdMarks,
				},
				{
					ID:        "complete-jellyfish-rows",
					Kind:      HintStepExplain,
					Message:   fourth,
					Narration: []HintNarrationChunk{narrationMessage(fourth)},
					Marks:     fourthMarks,
				},
				{
					ID:        "explain-jellyfish-columns",
					Kind:      HintStepExplain,
					Message:   consequence,
					Narration: []HintNarrationChunk{narrationMessage(consequence)},
					Marks:     consequenceMarks,
				},
				{
					ID:        "apply-jellyfish",
					Kind:      HintStepApply,
					Message:   apply,
					Narration: actionNarration(apply),
					Marks:     append(consequenceMarks, conclusionMarks(conclusion)...),
					Effect:    conclusionEffect(conclusion),
				},
			}
		}
		apply := hintDescriptor(
			"hint.fish-confinement.apply",
			digitArgument("digit", digit),
			cellsArgument("cells", positions),
			strategyArgument("strategy", move.Technique),
		)
		premiseMarks := premiseCandidateMarks(evidence)
		baseMarks := append([]HintMark(nil), premiseMarks...)
		for _, ref := range baseUnits {
			target, _ := unitTarget(&ref)
			baseMarks = append(baseMarks, HintMark{Target: target, Role: HintRoleFocus})
		}
		explainMarks := append([]HintMark(nil), baseMarks...)
		for _, ref := range coverUnits {
			target, _ := unitTarget(&ref)
			explainMarks = append(explainMarks, HintMark{Target: target, Role: HintRoleFocus})
		}
		return []HintStep{
			{
				ID:        "find-" + move.Technique,
				Kind:      HintStepObserve,
				Message:   observe,
				Narration: []HintNarrationChunk{narrationMessage(observe)},
				Marks:     baseMarks,
			},
			{
				ID:        "explain-" + move.Technique,
				Kind:      HintStepExplain,
				Message:   explain,
				Narration: []HintNarrationChunk{narrationMessage(explain)},
				Marks:     explainMarks,
			},
			{
				ID:        "apply-" + move.Technique,
				Kind:      conclusionKind(conclusion),
				Message:   apply,
				Narration: actionNarration(apply),
				Marks:     append(explainMarks, conclusionMarks(conclusion)...),
				Effect:    conclusionEffect(conclusion),
			},
		}
	}
	return []HintStep{
		{ID: "find-" + move.Technique, Kind: HintStepObserve, Marks: marks},
		{ID: "compare-" + move.Technique, Kind: HintStepExplain, Marks: marks},
		{
			ID:     "apply-" + move.Technique,
			Kind:   conclusionKind(conclusion),
			Marks:  append(marks, conclusionMarks(conclusion)...),
			Effect: conclusionEffect(conclusion),
		},
	}
}
