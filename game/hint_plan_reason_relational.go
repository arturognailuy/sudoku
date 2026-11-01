package game

import (
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/message"
	"github.com/gnailuy/sudoku/solver"
)

func relationalSteps(move *solver.Move, _ string) []HintStep {
	evidence := move.Evidence
	if evidence == nil || len(evidence.Premises) < 3 || len(evidence.Eliminations) == 0 {
		return evidenceSteps(move, HintConclusion{Placement: placement(move)})
	}
	conclusion := HintConclusion{Placement: placement(move), Eliminations: evidence.Eliminations}
	baseMarks := unitAndPremiseMarks(evidence)
	nodes := visualNodes(evidence, false)
	var firstNodes []HintNode
	var firstLinks, links []HintLink
	var observe, explain message.Descriptor
	actionDigit := 0
	switch move.Technique {
	case "xy-wing", "xyz-wing":
		pivot, wings := evidence.Premises[0], evidence.Premises[1:3]
		shared := sharedValues(wings...)
		if len(shared) != 1 || (move.Technique == "xy-wing" && len(evidence.Links) != 2) {
			return evidenceSteps(move, conclusion)
		}
		actionDigit = shared[0]
		wingCells := []core.Position{wings[0].Position, wings[1].Position}
		if move.Technique == "xy-wing" {
			if len(pivot.Values) != 2 || len(wings[0].Values) != 2 || len(wings[1].Values) != 2 {
				return evidenceSteps(move, conclusion)
			}
			wingForPivot := func(value int) (solver.CandidateGroup, bool) {
				for _, wing := range wings {
					if containsValue(wing.Values, value) {
						return wing, true
					}
				}
				return solver.CandidateGroup{}, false
			}
			firstWing, firstOK := wingForPivot(pivot.Values[0])
			secondWing, secondOK := wingForPivot(pivot.Values[1])
			if !firstOK || !secondOK || firstWing.Position == secondWing.Position {
				return evidenceSteps(move, conclusion)
			}
			pivotNodes := nodesForCandidateRefs(nodes,
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[0]},
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[1]},
			)
			firstCaseNodes := nodesForCandidateRefs(nodes,
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[0]},
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[1]},
				solver.CandidateRef{Position: firstWing.Position, Value: pivot.Values[0]},
				solver.CandidateRef{Position: firstWing.Position, Value: actionDigit},
			)
			firstCaseLink := HintLink{
				From: candidateNodeID(solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[0]}),
				To:   candidateNodeID(solver.CandidateRef{Position: firstWing.Position, Value: actionDigit}),
				Kind: solver.LinkImplication,
			}
			secondCaseLink := HintLink{
				From: candidateNodeID(solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[1]}),
				To:   candidateNodeID(solver.CandidateRef{Position: secondWing.Position, Value: actionDigit}),
				Kind: solver.LinkImplication,
			}
			caseLinks := []HintLink{firstCaseLink, secondCaseLink}
			pivotMessage := hintDescriptor(
				"hint.xy-wing.pivot",
				cellArgument("pivot", pivot.Position),
				digitsArgument("digits", pivot.Values),
			)
			firstCaseMessage := hintDescriptor(
				"hint.xy-wing.case",
				cellArgument("pivot", pivot.Position),
				digitArgument("pivot_digit", pivot.Values[0]),
				cellArgument("wing", firstWing.Position),
				digitArgument("shared_digit", actionDigit),
			)
			secondCaseMessage := hintDescriptor(
				"hint.xy-wing.case",
				cellArgument("pivot", pivot.Position),
				digitArgument("pivot_digit", pivot.Values[1]),
				cellArgument("wing", secondWing.Position),
				digitArgument("shared_digit", actionDigit),
			)
			guarantee := hintDescriptor(
				"hint.xy-wing.guarantee",
				cellArgument("pivot", pivot.Position),
				cellsArgument("wing_cells", wingCells),
				digitArgument("shared_digit", actionDigit),
			)
			var target core.Position
			if conclusion.Placement != nil {
				target = conclusion.Placement.Position
			} else if len(evidence.Eliminations) > 0 {
				target = evidence.Eliminations[0].Position
			} else {
				return evidenceSteps(move, conclusion)
			}
			targetRef := solver.CandidateRef{Position: target, Value: actionDigit}
			finalNodes := append([]HintNode(nil), nodes...)
			finalNodes = append(
				finalNodes,
				HintNode{ID: candidateNodeID(targetRef), Target: candidateTarget(targetRef)},
			)
			finalLinks := append([]HintLink(nil), caseLinks...)
			for _, wing := range wings {
				finalLinks = append(
					finalLinks,
					HintLink{
						From: candidateNodeID(solver.CandidateRef{Position: wing.Position, Value: actionDigit}),
						To:   candidateNodeID(targetRef),
						Kind: solver.LinkWeak,
					},
				)
			}
			var apply message.Descriptor
			if conclusion.Placement != nil {
				apply = hintDescriptor(
					"hint.xy-wing.place",
					cellArgument("cell", target),
					cellsArgument("wing_cells", wingCells),
					digitArgument("shared_digit", actionDigit),
					digitArgument("digit", conclusion.Placement.Value),
					strategyArgument("strategy", move.Technique),
				)
			} else {
				positions := make([]core.Position, len(evidence.Eliminations))
				for index, ref := range evidence.Eliminations {
					positions[index] = ref.Position
				}
				apply = hintDescriptor("hint.xy-wing.apply", cellsArgument("cells", positions), cellsArgument("wing_cells",
					wingCells), digitArgument("shared_digit", actionDigit), strategyArgument("strategy", move.Technique))
			}
			return []HintStep{
				{
					ID:        "find-xy-wing-pivot",
					Kind:      HintStepObserve,
					Message:   pivotMessage,
					Narration: []HintNarrationChunk{narrationMessage(pivotMessage)},
					Marks:     marksForNodes(pivotNodes),
					Nodes:     pivotNodes,
					Links:     []HintLink{},
				},
				{
					ID:        "follow-xy-wing-first-case",
					Kind:      HintStepExplain,
					Message:   firstCaseMessage,
					Narration: []HintNarrationChunk{narrationMessage(firstCaseMessage)},
					Marks:     marksForNodes(firstCaseNodes),
					Nodes:     firstCaseNodes,
					Links:     []HintLink{firstCaseLink},
				},
				{
					ID:        "follow-xy-wing-second-case",
					Kind:      HintStepExplain,
					Message:   secondCaseMessage,
					Narration: []HintNarrationChunk{narrationMessage(secondCaseMessage)},
					Marks:     marksForNodes(nodes),
					Nodes:     nodes,
					Links:     caseLinks,
				},
				{
					ID:        "combine-xy-wing-cases",
					Kind:      HintStepExplain,
					Message:   guarantee,
					Narration: []HintNarrationChunk{narrationMessage(guarantee)},
					Marks:     baseMarks,
					Nodes:     nodes,
					Links:     caseLinks,
				},
				{
					ID:        "apply-xy-wing",
					Kind:      conclusionKind(conclusion),
					Message:   apply,
					Narration: actionNarration(apply),
					Marks:     append(baseMarks, conclusionMarks(conclusion)...),
					Nodes:     finalNodes,
					Links:     finalLinks,
					Effect:    conclusionEffect(conclusion),
				},
			}
		}
		if move.Technique == "xyz-wing" {
			if len(pivot.Values) != 3 || len(wings[0].Values) != 2 || len(wings[1].Values) != 2 {
				return evidenceSteps(move, conclusion)
			}
			nonShared := make([]int, 0, 2)
			for _, value := range pivot.Values {
				if value != actionDigit {
					nonShared = append(nonShared, value)
				}
			}
			if len(nonShared) != 2 {
				return evidenceSteps(move, conclusion)
			}
			wingForPivot := func(value int) (solver.CandidateGroup, bool) {
				for _, wing := range wings {
					if containsValue(wing.Values, value) && containsValue(wing.Values, actionDigit) {
						return wing, true
					}
				}
				return solver.CandidateGroup{}, false
			}
			firstWing, firstOK := wingForPivot(nonShared[0])
			secondWing, secondOK := wingForPivot(nonShared[1])
			if !firstOK || !secondOK || firstWing.Position == secondWing.Position {
				return evidenceSteps(move, conclusion)
			}
			firstUnit, firstUnitOK := sharedHintUnit(pivot.Position, firstWing.Position)
			secondUnit, secondUnitOK := sharedHintUnit(pivot.Position, secondWing.Position)
			if !firstUnitOK || !secondUnitOK {
				return evidenceSteps(move, conclusion)
			}
			pivotNodes := nodesForCandidateRefs(nodes,
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[0]},
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[1]},
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[2]},
			)
			firstCaseNodes := nodesForCandidateRefs(nodes,
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[0]},
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[1]},
				solver.CandidateRef{Position: pivot.Position, Value: pivot.Values[2]},
				solver.CandidateRef{Position: firstWing.Position, Value: nonShared[0]},
				solver.CandidateRef{Position: firstWing.Position, Value: actionDigit},
			)
			firstCaseLink := HintLink{
				From: candidateNodeID(solver.CandidateRef{Position: pivot.Position, Value: nonShared[0]}),
				To:   candidateNodeID(solver.CandidateRef{Position: firstWing.Position, Value: actionDigit}),
				Kind: solver.LinkImplication,
			}
			secondCaseLink := HintLink{
				From: candidateNodeID(solver.CandidateRef{Position: pivot.Position, Value: nonShared[1]}),
				To:   candidateNodeID(solver.CandidateRef{Position: secondWing.Position, Value: actionDigit}),
				Kind: solver.LinkImplication,
			}
			caseLinks := []HintLink{firstCaseLink, secondCaseLink}
			pivotMessage := hintDescriptor(
				"hint.xyz-wing.pivot",
				cellArgument("pivot", pivot.Position),
				digitsArgument("digits", pivot.Values),
			)
			firstCaseMessage := hintDescriptor(
				"hint.xyz-wing.case",
				cellArgument("pivot", pivot.Position),
				digitArgument("pivot_digit", nonShared[0]),
				cellArgument("wing", firstWing.Position),
				unitsArgument("unit", []solver.UnitRef{firstUnit}),
				digitArgument("shared_digit", actionDigit),
			)
			secondCaseMessage := hintDescriptor(
				"hint.xyz-wing.case",
				cellArgument("pivot", pivot.Position),
				digitArgument("pivot_digit", nonShared[1]),
				cellArgument("wing", secondWing.Position),
				unitsArgument("unit", []solver.UnitRef{secondUnit}),
				digitArgument("shared_digit", actionDigit),
			)
			pivotCaseMessage := hintDescriptor(
				"hint.xyz-wing.pivot-case",
				cellArgument("pivot", pivot.Position),
				digitArgument("shared_digit", actionDigit),
			)
			patternCells := []core.Position{pivot.Position, firstWing.Position, secondWing.Position}
			guarantee := hintDescriptor(
				"hint.xyz-wing.guarantee",
				cellArgument("pivot", pivot.Position),
				cellsArgument("pattern_cells", patternCells),
				digitArgument("shared_digit", actionDigit),
			)
			target := evidence.Eliminations[0].Position
			targetRef := solver.CandidateRef{Position: target, Value: actionDigit}
			finalNodes := append([]HintNode(nil), nodes...)
			finalNodes = append(
				finalNodes,
				HintNode{ID: candidateNodeID(targetRef), Target: candidateTarget(targetRef)},
			)
			finalLinks := append([]HintLink(nil), caseLinks...)
			witnessUnits := make([]solver.UnitRef, 0, len(patternCells))
			for _, patternCell := range patternCells {
				unit, ok := sharedHintUnit(target, patternCell)
				if !ok {
					return evidenceSteps(move, conclusion)
				}
				witnessUnits = append(witnessUnits, unit)
				finalLinks = append(
					finalLinks,
					HintLink{
						From: candidateNodeID(solver.CandidateRef{Position: patternCell, Value: actionDigit}),
						To:   candidateNodeID(targetRef),
						Kind: solver.LinkWeak,
					},
				)
			}
			messagePatternCells := make([]core.Position, 0, len(patternCells))
			messageWitnessUnits := make([]solver.UnitRef, 0, len(witnessUnits))
			for _, preferredKind := range []solver.UnitKind{solver.UnitRow, solver.UnitColumn, solver.UnitBox} {
				for index, unit := range witnessUnits {
					if unit.Kind == preferredKind {
						messagePatternCells = append(messagePatternCells, patternCells[index])
						messageWitnessUnits = append(messageWitnessUnits, unit)
					}
				}
			}
			if len(messagePatternCells) != 3 {
				return evidenceSteps(move, conclusion)
			}
			var apply message.Descriptor
			if conclusion.Placement != nil {
				apply = hintDescriptor(
					"hint.relational-wing.place",
					digitArgument("removed_digit", actionDigit),
					digitArgument("digit", conclusion.Placement.Value),
					cellArgument("cell", conclusion.Placement.Position),
					strategyArgument("strategy", move.Technique),
				)
			} else {
				apply = hintDescriptor("hint.xyz-wing.apply",
					cellArgument("target", target),
					cellArgument("first_pattern_cell", messagePatternCells[0]), unitsArgument("first_unit", []solver.UnitRef{messageWitnessUnits[0]}),
					cellArgument("second_pattern_cell", messagePatternCells[1]), unitsArgument("second_unit", []solver.UnitRef{messageWitnessUnits[1]}),
					cellArgument("third_pattern_cell", messagePatternCells[2]), unitsArgument("third_unit", []solver.UnitRef{messageWitnessUnits[2]}),
					digitArgument("shared_digit", actionDigit), strategyArgument("strategy", move.Technique))
			}
			return []HintStep{
				{
					ID:        "find-xyz-wing-pivot",
					Kind:      HintStepObserve,
					Message:   pivotMessage,
					Narration: []HintNarrationChunk{narrationMessage(pivotMessage)},
					Marks:     marksForNodes(pivotNodes),
					Nodes:     pivotNodes,
					Links:     []HintLink{},
				},
				{
					ID:        "follow-xyz-wing-first-case",
					Kind:      HintStepExplain,
					Message:   firstCaseMessage,
					Narration: []HintNarrationChunk{narrationMessage(firstCaseMessage)},
					Marks:     marksForNodes(firstCaseNodes),
					Nodes:     firstCaseNodes,
					Links:     []HintLink{firstCaseLink},
				},
				{
					ID:        "follow-xyz-wing-second-case",
					Kind:      HintStepExplain,
					Message:   secondCaseMessage,
					Narration: []HintNarrationChunk{narrationMessage(secondCaseMessage)},
					Marks:     marksForNodes(nodes),
					Nodes:     nodes,
					Links:     caseLinks,
				},
				{
					ID:        "cover-xyz-wing-shared-case",
					Kind:      HintStepExplain,
					Message:   pivotCaseMessage,
					Narration: []HintNarrationChunk{narrationMessage(pivotCaseMessage)},
					Marks:     marksForNodes(nodes),
					Nodes:     nodes,
					Links:     caseLinks,
				},
				{
					ID:        "combine-xyz-wing-cases",
					Kind:      HintStepExplain,
					Message:   guarantee,
					Narration: []HintNarrationChunk{narrationMessage(guarantee)},
					Marks:     baseMarks,
					Nodes:     nodes,
					Links:     caseLinks,
				},
				{
					ID:        "apply-xyz-wing",
					Kind:      conclusionKind(conclusion),
					Message:   apply,
					Narration: actionNarration(apply),
					Marks:     append(baseMarks, conclusionMarks(conclusion)...),
					Nodes:     finalNodes,
					Links:     finalLinks,
					Effect:    conclusionEffect(conclusion),
				},
			}
		}
		observe = hintDescriptor(
			"hint.pivot-wing.observe",
			cellArgument("pivot", pivot.Position),
			digitsArgument("pivot_digits", pivot.Values),
			cellsArgument("wing_cells", wingCells),
		)
		explain = hintDescriptor(
			"hint.pivot-wing.explain",
			digitArgument("shared_digit", actionDigit),
			cellArgument("pivot", pivot.Position),
			cellsArgument("wing_cells", wingCells),
		)
		links = visualLinks(evidence)
		if len(links) > 0 {
			firstNodes, firstLinks = visualPrefix(nodes, links)
		} else {
			for _, node := range nodes {
				if node.Target.Position != nil && *node.Target.Position == pivot.Position {
					firstNodes = append(firstNodes, node)
				}
			}
		}
	case "w-wing":
		if len(evidence.Premises) < 4 || len(evidence.Links) != 1 || evidence.Links[0].Kind != solver.LinkStrong {
			return evidenceSteps(move, conclusion)
		}
		wings := evidence.Premises[:2]
		pair := sharedValues(wings...)
		linkDigit := evidence.Links[0].From.Value
		if len(pair) != 2 || evidence.Links[0].To.Value != linkDigit {
			return evidenceSteps(move, conclusion)
		}
		for _, digit := range pair {
			if digit != linkDigit {
				actionDigit = digit
			}
		}
		wingCells := []core.Position{wings[0].Position, wings[1].Position}
		linkCells := []core.Position{evidence.Links[0].From.Position, evidence.Links[0].To.Position}
		linkUnit, ok := sharedHintUnit(linkCells[0], linkCells[1])
		if !ok {
			return evidenceSteps(move, conclusion)
		}
		observe = hintDescriptor(
			"hint.strong-link-wing.observe",
			cellsArgument("wing_cells", wingCells),
			digitsArgument("digits", pair),
		)
		bridge := hintDescriptor(
			"hint.strong-link-wing.bridge",
			unitsArgument("unit", []solver.UnitRef{linkUnit}),
			digitArgument("link_digit", linkDigit),
			cellsArgument("link_cells", linkCells),
		)
		force := hintDescriptor("hint.strong-link-wing.force",
			digitArgument("link_digit", linkDigit),
			cellArgument("first_link_cell", linkCells[0]),
			cellArgument("first_wing_cell", wingCells[0]),
			cellArgument("second_link_cell", linkCells[1]),
			cellArgument("second_wing_cell", wingCells[1]),
			digitArgument("forced_digit", actionDigit),
		)
		links = visualLinks(evidence)
		caseLinks := append([]HintLink(nil), links...)
		caseLinks = append(
			caseLinks,
			HintLink{
				From: candidateNodeID(solver.CandidateRef{Position: linkCells[0], Value: linkDigit}),
				To:   candidateNodeID(solver.CandidateRef{Position: wingCells[0], Value: actionDigit}),
				Kind: solver.LinkImplication,
			},
			HintLink{
				From: candidateNodeID(solver.CandidateRef{Position: linkCells[1], Value: linkDigit}),
				To:   candidateNodeID(solver.CandidateRef{Position: wingCells[1], Value: actionDigit}),
				Kind: solver.LinkImplication,
			},
		)
		bridgeMarks := marksForNodes(nodes)
		if target, valid := unitTarget(&linkUnit); valid {
			bridgeMarks = append(bridgeMarks, HintMark{Target: target, Role: HintRoleFocus})
		}
		for _, node := range nodes {
			if node.Target.Position != nil &&
				(*node.Target.Position == wingCells[0] || *node.Target.Position == wingCells[1]) {
				firstNodes = append(firstNodes, node)
			}
		}
		var apply message.Descriptor
		if conclusion.Placement != nil {
			apply = hintDescriptor(
				"hint.strong-link-wing.place",
				digitArgument("removed_digit", actionDigit),
				digitArgument("digit", conclusion.Placement.Value),
				cellArgument("cell", conclusion.Placement.Position),
				cellsArgument("wing_cells", wingCells),
				strategyArgument("strategy", move.Technique),
			)
		} else {
			positions := make([]core.Position, len(evidence.Eliminations))
			for i, ref := range evidence.Eliminations {
				positions[i] = ref.Position
			}
			apply = hintDescriptor("hint.strong-link-wing.apply", digitArgument("digit", actionDigit),
				cellsArgument("cells", positions), cellsArgument("wing_cells", wingCells), strategyArgument("strategy", move.Technique))
		}
		return []HintStep{
			{
				ID:        "find-w-wing-pair",
				Kind:      HintStepObserve,
				Message:   observe,
				Narration: []HintNarrationChunk{narrationMessage(observe)},
				Marks:     marksForNodes(firstNodes),
				Nodes:     firstNodes,
				Links:     []HintLink{},
			},
			{
				ID:        "find-w-wing-bridge",
				Kind:      HintStepObserve,
				Message:   bridge,
				Narration: []HintNarrationChunk{narrationMessage(bridge)},
				Marks:     bridgeMarks,
				Nodes:     nodes,
				Links:     links,
			},
			{
				ID:        "explain-w-wing-cases",
				Kind:      HintStepExplain,
				Message:   force,
				Narration: []HintNarrationChunk{narrationMessage(force)},
				Marks:     bridgeMarks,
				Nodes:     nodes,
				Links:     caseLinks,
			},
			{
				ID:        "apply-w-wing",
				Kind:      conclusionKind(conclusion),
				Message:   apply,
				Narration: actionNarration(apply),
				Marks:     append(bridgeMarks, conclusionMarks(conclusion)...),
				Nodes:     nodes,
				Links:     caseLinks,
				Effect:    conclusionEffect(conclusion),
			},
		}
	default:
		return evidenceSteps(move, conclusion)
	}
	if actionDigit == 0 {
		return evidenceSteps(move, conclusion)
	}
	var apply message.Descriptor
	if conclusion.Placement != nil {
		apply = hintDescriptor(
			"hint.relational-wing.place",
			digitArgument("removed_digit", actionDigit),
			digitArgument("digit", conclusion.Placement.Value),
			cellArgument("cell", conclusion.Placement.Position),
			strategyArgument("strategy", move.Technique),
		)
	} else {
		positions := make([]core.Position, len(evidence.Eliminations))
		for i, ref := range evidence.Eliminations {
			positions[i] = ref.Position
		}
		apply = hintDescriptor("hint.relational-wing.apply", digitArgument("digit", actionDigit), cellsArgument("cells",
			positions), strategyArgument("strategy", move.Technique))
	}
	return []HintStep{
		{
			ID:        "find-" + move.Technique,
			Kind:      HintStepObserve,
			Message:   observe,
			Narration: []HintNarrationChunk{narrationMessage(observe)},
			Marks:     marksForNodes(firstNodes),
			Nodes:     firstNodes,
			Links:     firstLinks,
		},
		{
			ID:        "explain-" + move.Technique,
			Kind:      HintStepExplain,
			Message:   explain,
			Narration: []HintNarrationChunk{narrationMessage(explain)},
			Marks:     baseMarks,
			Nodes:     nodes,
			Links:     links,
		},
		{
			ID:        "apply-" + move.Technique,
			Kind:      conclusionKind(conclusion),
			Message:   apply,
			Narration: actionNarration(apply),
			Marks:     append(baseMarks, conclusionMarks(conclusion)...),
			Nodes:     nodes,
			Links:     links,
			Effect:    conclusionEffect(conclusion),
		},
	}
}
