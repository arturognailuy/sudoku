package game

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

func TestFishAndRelationalFamiliesTeachBeforeNaming(t *testing.T) {
	games := realGamesForHintStrategies(t)
	store := solver.NewStore()
	keys := store.GetAllStrategySolverKeys()
	families := []string{"x-wing", "swordfish", "jellyfish", "xy-wing", "w-wing", "xyz-wing"}
	for _, key := range families {
		t.Run(key, func(t *testing.T) {
			options := NewDefaultOptions(store)
			options.StrategySolverKeys = keys
			current := NewGame(boardFromPuzzle(t, games[key]), options)
			var plan *HintPlan
			for number := 0; number < 128; number++ {
				plan = current.Hint()
				if plan == nil {
					t.Fatalf("exhausted hints before %s", key)
				}
				if plan.Strategy.ID == key {
					break
				}
				if _, err := current.Apply(ApplyHint{PlanID: plan.PlanID}); err != nil {
					t.Fatal(err)
				}
			}
			expectedSteps := 3
			if key == "x-wing" {
				expectedSteps = 4
			}
			if key == "swordfish" {
				expectedSteps = 5
			}
			if key == "jellyfish" {
				expectedSteps = 6
			}
			if key == "xy-wing" {
				expectedSteps = 5
			}
			if key == "xyz-wing" {
				expectedSteps = 6
			}
			if key == "w-wing" {
				expectedSteps = 4
			}
			if plan == nil || plan.Strategy.ID != key || len(plan.Steps) != expectedSteps {
				t.Fatalf("unexpected family plan: %+v", plan)
			}
			for _, step := range plan.Steps[:len(plan.Steps)-1] {
				if strings.Contains(step.Message.Text, plan.Strategy.Name.Text) {
					t.Fatalf("strategy vocabulary appears before the reasoning: %q", step.Message.Text)
				}
				if strings.Contains(step.Message.Text, "(") {
					t.Fatalf("family plan exposes coordinate shorthand: %q", step.Message.Text)
				}
			}
			if key == "x-wing" || key == "swordfish" || key == "jellyfish" {
				if plan.Reason == nil || plan.Reason.Kind != HintReasonFishConfinement {
					t.Fatalf("fish contract = reason %+v", plan.Reason)
				}
				expectedSize := map[string]int{"x-wing": 2, "swordfish": 3, "jellyfish": 4}[key]
				reason := plan.Reason.FishConfinement
				if reason == nil || len(reason.BaseUnits) != expectedSize || len(reason.CoverUnits) != expectedSize ||
					len(reason.PatternCandidates) < expectedSize ||
					len(reason.ActionTargets) == 0 {
					t.Fatalf("fish reason is incomplete: %+v", reason)
				}
				if key == "x-wing" {
					wantMessages := []string{
						"hint.x-wing.first",
						"hint.x-wing.second",
						"hint.x-wing.consequence",
						"hint.x-wing.apply",
					}
					for index, want := range wantMessages {
						if plan.Steps[index].Message.ID != want {
							t.Fatalf("x-wing step %d message = %q, want %q", index, plan.Steps[index].Message.ID, want)
						}
					}
					if got := []int{len(plan.Steps[0].Marks), len(plan.Steps[1].Marks), len(plan.Steps[2].Marks),
						len(plan.Steps[3].Marks)}; !reflect.DeepEqual(
						got,
						[]int{2, 6, 8, 10},
					) {
						t.Fatalf("x-wing mark progression = %v, want [2 6 8 10]", got)
					}
					if got := plan.Steps[0].Message.Text; got != "In row 5, candidate 7 appears only in column 1 and column 4." {
						t.Fatalf("x-wing first-step copy = %q", got)
					}
					wantEliminations := []solver.CandidateRef{
						{Position: core.NewPosition(0, 3), Value: 7},
						{Position: core.NewPosition(7, 3), Value: 7},
					}
					if !reflect.DeepEqual(plan.Conclusion.Eliminations, wantEliminations) {
						t.Fatalf("x-wing eliminations = %+v, want %+v", plan.Conclusion.Eliminations, wantEliminations)
					}
				} else if key == "swordfish" {
					wantMessages := []string{"hint.fish-confinement.observe", "hint.fish-confinement.observe",
						"hint.fish-confinement.observe", "hint.swordfish.consequence", "hint.fish-confinement.apply"}
					for index, want := range wantMessages {
						if plan.Steps[index].Message.ID != want {
							t.Fatalf("swordfish step %d message = %q, want %q", index, plan.Steps[index].Message.ID, want)
						}
					}
					if got := []int{len(plan.Steps[0].Marks), len(plan.Steps[1].Marks), len(plan.Steps[2].Marks),
						len(plan.Steps[3].Marks), len(plan.Steps[4].Marks)}; !reflect.DeepEqual(got, []int{2, 6, 9, 12, 13}) {
						t.Fatalf("swordfish mark progression = %v, want [2 6 9 12 13]", got)
					}
					wantCopy := []string{
						"In row 3, candidate 3 appears only in column 2 and column 6.",
						"In row 8, candidate 3 appears only in column 2 and column 9.",
						"In row 9, candidate 3 appears only in column 6 and column 9.",
					}
					for index, want := range wantCopy {
						if plan.Steps[index].Message.Text != want {
							t.Fatalf("swordfish step %d copy = %q, want %q", index, plan.Steps[index].Message.Text, want)
						}
					}
				} else if key == "jellyfish" {
					wantMessages := []string{"hint.fish-confinement.observe", "hint.fish-confinement.observe",
						"hint.fish-confinement.observe", "hint.fish-confinement.observe", "hint.jellyfish.consequence", "hint.fish-confinement.apply"}
					for index, want := range wantMessages {
						if plan.Steps[index].Message.ID != want {
							t.Fatalf("jellyfish step %d message = %q, want %q", index, plan.Steps[index].Message.ID, want)
						}
					}
					if got := []int{len(plan.Steps[0].Marks), len(plan.Steps[1].Marks), len(plan.Steps[2].Marks),
						len(plan.Steps[3].Marks), len(plan.Steps[4].Marks), len(plan.Steps[5].Marks)}; !reflect.DeepEqual(got, []int{2, 6, 9, 12, 16, 21}) {
						t.Fatalf("jellyfish mark progression = %v, want [2 6 9 12 16 21]", got)
					}
					wantCopy := []string{
						"In row 2, candidate 9 appears only in column 1 and column 4.",
						"In row 3, candidate 9 appears only in column 1 and column 9.",
						"In row 4, candidate 9 appears only in column 3 and column 4.",
						"In row 7, candidate 9 appears only in column 3 and column 9.",
					}
					for index, want := range wantCopy {
						if plan.Steps[index].Message.Text != want {
							t.Fatalf("jellyfish step %d copy = %q, want %q", index, plan.Steps[index].Message.Text, want)
						}
					}
					wantEliminations := []solver.CandidateRef{
						{Position: core.NewPosition(0, 3), Value: 9},
						{Position: core.NewPosition(4, 3), Value: 9},
						{Position: core.NewPosition(7, 3), Value: 9},
						{Position: core.NewPosition(8, 3), Value: 9},
						{Position: core.NewPosition(8, 8), Value: 9},
					}
					if !reflect.DeepEqual(plan.Conclusion.Eliminations, wantEliminations) {
						t.Fatalf("jellyfish eliminations = %+v, want %+v", plan.Conclusion.Eliminations, wantEliminations)
					}
				} else if plan.Steps[2].Message.ID != "hint.fish-confinement.apply" {
					t.Fatalf("fish conclusion = %+v", plan.Steps[2].Message)
				}
			} else {
				if strings.HasPrefix(plan.Steps[1].Message.ID, "hint.strategy.") ||
					strings.HasPrefix(plan.Steps[len(plan.Steps)-1].Message.ID, "hint.strategy.") {
					t.Fatalf("relational-wing descriptors remain generic: %+v", plan.Steps)
				}
				expectedApply := "hint.relational-wing.apply"
				if plan.Conclusion.Placement != nil {
					expectedApply = "hint.relational-wing.place"
				}
				if key == "xy-wing" {
					expectedApply = "hint.xy-wing.apply"
					if plan.Conclusion.Placement != nil {
						expectedApply = "hint.xy-wing.place"
					}
				}
				if key == "xyz-wing" {
					expectedApply = "hint.xyz-wing.apply"
				}
				if key == "w-wing" {
					expectedApply = "hint.strong-link-wing.apply"
					if plan.Conclusion.Placement != nil {
						expectedApply = "hint.strong-link-wing.place"
					}
				}
				if plan.Steps[len(plan.Steps)-1].Message.ID != expectedApply {
					t.Fatalf("relational-wing action descriptor = %q, want %q", plan.Steps[len(plan.Steps)-1].Message.ID, expectedApply)
				}
			}
			switch key {
			case "xy-wing":
				wantMessages := []string{
					"hint.xy-wing.pivot",
					"hint.xy-wing.case",
					"hint.xy-wing.case",
					"hint.xy-wing.guarantee",
					"hint.xy-wing.place",
				}
				wantLinks := []int{0, 1, 2, 2, 4}
				for index, step := range plan.Steps {
					if step.Message.ID != wantMessages[index] || len(step.Nodes) == 0 ||
						len(step.Links) != wantLinks[index] {
						t.Fatalf("XY-Wing step %d is incomplete: %+v", index, step)
					}
				}
				for _, step := range plan.Steps[1:4] {
					for _, link := range step.Links {
						if link.Kind != solver.LinkImplication {
							t.Fatalf("XY-Wing case step has non-implication link: %+v", step)
						}
					}
				}
				if plan.Steps[4].Links[2].Kind != solver.LinkWeak || plan.Steps[4].Links[3].Kind != solver.LinkWeak {
					t.Fatalf("XY-Wing result lacks both target witnesses: %+v", plan.Steps[4])
				}
			case "xyz-wing":
				wantMessages := []string{
					"hint.xyz-wing.pivot",
					"hint.xyz-wing.case",
					"hint.xyz-wing.case",
					"hint.xyz-wing.pivot-case",
					"hint.xyz-wing.guarantee",
					"hint.xyz-wing.apply",
				}
				wantLinks := []int{0, 1, 2, 2, 2, 5}
				for index, step := range plan.Steps {
					if step.Message.ID != wantMessages[index] || len(step.Nodes) == 0 ||
						len(step.Links) != wantLinks[index] {
						t.Fatalf("XYZ-Wing step %d is incomplete: %+v", index, step)
					}
				}
				for _, step := range plan.Steps[1:5] {
					for _, link := range step.Links {
						if link.Kind != solver.LinkImplication {
							t.Fatalf("XYZ-Wing case step has non-implication link: %+v", step)
						}
					}
				}
				for _, link := range plan.Steps[5].Links[2:] {
					if link.Kind != solver.LinkWeak {
						t.Fatalf("XYZ-Wing result lacks all three target witnesses: %+v", plan.Steps[5])
					}
				}
			case "w-wing":
				if len(plan.Steps[0].Links) != 0 {
					t.Fatalf("W-Wing introduces the bridge before the paired cells: %+v", plan.Steps[0])
				}
				if plan.Steps[1].Message.ID != "hint.strong-link-wing.bridge" ||
					plan.Steps[2].Message.ID != "hint.strong-link-wing.force" {
					t.Fatalf("W-Wing causal descriptors are incomplete: %+v", plan.Steps)
				}
				if len(plan.Steps[1].Links) != 1 || plan.Steps[1].Links[0].Kind != solver.LinkStrong {
					t.Fatalf("W-Wing bridge step is incomplete: %+v", plan.Steps[1])
				}
				for _, step := range plan.Steps[2:] {
					if len(step.Nodes) == 0 || len(step.Links) != 3 || step.Links[0].Kind != solver.LinkStrong ||
						step.Links[1].Kind != solver.LinkImplication ||
						step.Links[2].Kind != solver.LinkImplication {
						t.Fatalf("W-Wing case implications are incomplete: %+v", step)
					}
				}
			}
		})
	}
}

func TestUniquenessFamilyTeachesBeforeNaming(t *testing.T) {
	games := realGamesForHintStrategies(t)
	store := solver.NewStore()
	keys := store.GetAllStrategySolverKeys()
	families := []string{
		"bug-plus-one",
		"unique-rectangle",
		"unique-rectangle-2",
		"unique-rectangle-3",
		"unique-rectangle-4",
	}
	for _, key := range families {
		t.Run(key, func(t *testing.T) {
			options := NewDefaultOptions(store)
			options.StrategySolverKeys = keys
			current := NewGame(boardFromPuzzle(t, games[key]), options)
			var plan *HintPlan
			for number := 0; number < 128; number++ {
				plan = current.Hint()
				if plan == nil {
					t.Fatalf("exhausted hints before %s", key)
				}
				if plan.Strategy.ID == key {
					break
				}
				if _, err := current.Apply(ApplyHint{PlanID: plan.PlanID}); err != nil {
					t.Fatal(err)
				}
			}
			expectedSteps := 3
			if key == "bug-plus-one" || key == "unique-rectangle" || key == "unique-rectangle-2" ||
				key == "unique-rectangle-4" {
				expectedSteps = 6
			}
			if key == "unique-rectangle-3" {
				expectedSteps = 7
			}
			if plan == nil || plan.Strategy.ID != key || len(plan.Steps) != expectedSteps {
				t.Fatalf("unexpected family plan: %+v", plan)
			}
			for _, step := range plan.Steps[:len(plan.Steps)-1] {
				if strings.Contains(step.Message.Text, plan.Strategy.Name.Text) {
					t.Fatalf("strategy vocabulary appears before the reasoning: %q", step.Message.Text)
				}
				if strings.Contains(step.Message.Text, "(") {
					t.Fatalf("family plan exposes coordinate shorthand: %q", step.Message.Text)
				}
			}
			if key == "bug-plus-one" {
				ids := []string{
					"hint.bug-plus-one.bivalue",
					"hint.bug-plus-one.trivalue",
					"hint.bug-plus-one.unit-count",
					"hint.bug-plus-one.unit-count",
					"hint.bug-plus-one.explain",
					"hint.bug-plus-one.place",
				}
				for index, id := range ids {
					if plan.Steps[index].Message.ID != id {
						t.Fatalf("BUG+1 descriptor %d = %q, want %q", index, plan.Steps[index].Message.ID, id)
					}
				}
				markCounts := []int{12, 16, 17, 18, 19, 20}
				for index, count := range markCounts {
					if len(plan.Steps[index].Marks) != count {
						t.Fatalf("BUG+1 mark count %d = %d, want %d", index, len(plan.Steps[index].Marks), count)
					}
				}
			} else if key == "unique-rectangle" {
				ids := []string{"hint.unique-rectangle.rectangle", "hint.unique-rectangle.pair-corners",
					"hint.unique-rectangle.third-corner", "hint.unique-rectangle.expanded-corner",
					"hint.unique-rectangle.protect-uniqueness", "hint.unique-rectangle.type1-place"}
				markCounts := []int{8, 12, 14, 17, 17, 18}
				for index, id := range ids {
					if plan.Steps[index].Message.ID != id || len(plan.Steps[index].Marks) != markCounts[index] {
						t.Fatalf("Unique Rectangle Type 1 step %d = %+v", index, plan.Steps[index])
					}
				}
			} else if key == "unique-rectangle-2" {
				ids := []string{"hint.unique-rectangle.rectangle", "hint.unique-rectangle.pair-corners",
					"hint.unique-rectangle.type2-first-expanded", "hint.unique-rectangle.type2-shared-extra",
					"hint.unique-rectangle.type2-protect-uniqueness", "hint.unique-rectangle.type2-place"}
				markCounts := []int{8, 12, 15, 18, 19, 20}
				for index, id := range ids {
					if plan.Steps[index].Message.ID != id || len(plan.Steps[index].Marks) != markCounts[index] {
						t.Fatalf("Unique Rectangle Type 2 step %d = %+v", index, plan.Steps[index])
					}
				}
			} else if key == "unique-rectangle-3" {
				ids := []string{"hint.unique-rectangle.rectangle", "hint.unique-rectangle.pair-corners",
					"hint.unique-rectangle.type3-first-expanded", "hint.unique-rectangle.type3-second-expanded",
					"hint.unique-rectangle.type3-partner", "hint.unique-rectangle.type3-reserve", "hint.unique-rectangle.type3-place"}
				markCounts := []int{8, 12, 15, 18, 20, 21, 22}
				for index, id := range ids {
					if plan.Steps[index].Message.ID != id || len(plan.Steps[index].Marks) != markCounts[index] {
						t.Fatalf("Unique Rectangle Type 3 step %d = %+v", index, plan.Steps[index])
					}
				}
			} else if key == "unique-rectangle-4" {
				ids := []string{"hint.unique-rectangle.rectangle", "hint.unique-rectangle.pair-corners",
					"hint.unique-rectangle.type4-first-expanded", "hint.unique-rectangle.type4-second-expanded",
					"hint.unique-rectangle.type4", "hint.unique-rectangle.type4-apply"}
				markCounts := []int{8, 12, 15, 18, 20, 20}
				linkCounts := []int{0, 0, 0, 0, 1, 1}
				for index, id := range ids {
					if plan.Steps[index].Message.ID != id || len(plan.Steps[index].Marks) != markCounts[index] ||
						len(plan.Steps[index].Links) != linkCounts[index] {
						t.Fatalf("Unique Rectangle Type 4 step %d = %+v", index, plan.Steps[index])
					}
				}
				for _, step := range plan.Steps[4:] {
					if len(step.Nodes) != 2 || step.Links[0].Kind != solver.LinkStrong {
						t.Fatalf("Unique Rectangle Type 4 conjugate pair is incomplete: %+v", step)
					}
				}
			} else {
				explainID := map[string]string{
					"unique-rectangle":   "hint.unique-rectangle.type1",
					"unique-rectangle-2": "hint.unique-rectangle.type2",
					"unique-rectangle-3": "hint.unique-rectangle.type3",
					"unique-rectangle-4": "hint.unique-rectangle.type4",
				}[key]
				if plan.Steps[0].Message.ID != "hint.unique-rectangle.observe" || plan.Steps[1].Message.ID != explainID {
					t.Fatalf("unique-rectangle descriptors remain generic: %+v", plan.Steps)
				}
				expectedApply := "hint.unique-rectangle.apply"
				if plan.Conclusion.Placement != nil {
					expectedApply = "hint.unique-rectangle.place"
				}
				if plan.Steps[2].Message.ID != expectedApply {
					t.Fatalf("unique-rectangle action descriptor = %q, want %q", plan.Steps[2].Message.ID, expectedApply)
				}
			}
		})
	}
}

func TestColoringAndChainFamiliesTeachBeforeNaming(t *testing.T) {
	games := realGamesForHintStrategies(t)
	store := solver.NewStore()
	keys := store.GetAllStrategySolverKeys()
	for _, key := range []string{"simple-coloring", "x-cycles", "xy-chain"} {
		t.Run(key, func(t *testing.T) {
			options := NewDefaultOptions(store)
			options.StrategySolverKeys = keys
			current := NewGame(boardFromPuzzle(t, games[key]), options)
			var plan *HintPlan
			for number := 0; number < 128; number++ {
				plan = current.Hint()
				if plan == nil {
					t.Fatalf("exhausted hints before %s", key)
				}
				if plan.Strategy.ID == key {
					break
				}
				if _, err := current.Apply(ApplyHint{PlanID: plan.PlanID}); err != nil {
					t.Fatal(err)
				}
			}
			if plan == nil || plan.Strategy.ID != key {
				t.Fatalf("unexpected family plan: %+v", plan)
			}
			switch key {
			case "xy-chain":
				if plan.Reason == nil || plan.Reason.ImplicationChain == nil || len(plan.Steps) != 6 {
					t.Fatalf("unexpected logic-sized XY-chain plan: %+v", plan)
				}
			case "simple-coloring":
				if len(plan.Steps) != 6 {
					t.Fatalf("Simple Coloring should mirror the approved six-step progression: %+v", plan)
				}
			case "x-cycles":
				if len(plan.Steps) != 5 {
					t.Fatalf("X-Cycles should mirror the approved five-step progression: %+v", plan)
				}
			default:
				if len(plan.Steps) != 3 {
					t.Fatalf("unexpected family plan: %+v", plan)
				}
			}
			preVocabularySteps := plan.Steps[:2]
			if key == "simple-coloring" || key == "x-cycles" || key == "xy-chain" {
				preVocabularySteps = plan.Steps[:len(plan.Steps)-1]
			}
			for _, step := range preVocabularySteps {
				if strings.Contains(step.Message.Text, plan.Strategy.Name.Text) {
					t.Fatalf("strategy vocabulary appears before the reasoning: %q", step.Message.Text)
				}
				if strings.Contains(step.Message.Text, "r1c") || strings.Contains(step.Message.Text, "(") {
					t.Fatalf("family plan exposes compact notation: %q", step.Message.Text)
				}
			}
			last := plan.Steps[len(plan.Steps)-1]
			switch key {
			case "xy-chain":
				if plan.Reason == nil || plan.Reason.Kind != HintReasonImplicationChain ||
					last.Message.ID != "hint.xy-chain.place" {
					t.Fatalf("xy-chain pilot contract = reason %+v, conclusion %+v", plan.Reason, last.Message)
				}
				want := []string{
					"hint.xy-chain.start",
					"hint.xy-chain.assume",
					"hint.xy-chain.follow",
					"hint.xy-chain.follow",
					"hint.xy-chain.reach",
					"hint.xy-chain.place",
				}
				for index, id := range want {
					if plan.Steps[index].Message.ID != id {
						t.Fatalf("XY-Chain step %d descriptor = %q, want %q", index+1, plan.Steps[index].Message.ID, id)
					}
				}
				for index, linkCount := range []int{0, 2, 4, 6, 7, 7} {
					if len(plan.Steps[index].Links) != linkCount {
						t.Fatalf(
							"XY-Chain step %d links = %d, want %d",
							index+1,
							len(plan.Steps[index].Links),
							linkCount,
						)
					}
				}
			case "simple-coloring":
				want := []string{
					"hint.coloring.pair",
					"hint.coloring.chain",
					"hint.coloring.alternatives",
					"hint.coloring.elimination",
					"hint.coloring.elimination",
					"hint.coloring.apply",
				}
				for index, id := range want {
					if plan.Steps[index].Message.ID != id {
						t.Fatalf(
							"Simple Coloring step %d descriptor = %q, want %q",
							index+1,
							plan.Steps[index].Message.ID,
							id,
						)
					}
				}
			case "x-cycles":
				want := []string{
					"hint.x-cycle.opening",
					"hint.x-cycle.assume",
					"hint.x-cycle.follow",
					"hint.x-cycle.close",
					"hint.x-cycle.place",
				}
				for index, id := range want {
					if plan.Steps[index].Message.ID != id {
						t.Fatalf("X-Cycles step %d descriptor = %q, want %q", index+1, plan.Steps[index].Message.ID, id)
					}
				}
				for index, linkCount := range []int{1, 3, 5, 7, 7} {
					if len(plan.Steps[index].Links) != linkCount {
						t.Fatalf(
							"X-Cycles step %d links = %d, want %d",
							index+1,
							len(plan.Steps[index].Links),
							linkCount,
						)
					}
				}
			}
			if len(plan.Steps[1].Links) <= len(plan.Steps[0].Links) {
				t.Fatalf("complex scene is not progressive: %+v", plan.Steps)
			}
			for _, mark := range plan.Steps[1].Marks {
				if mark.Target.Kind == HintTargetRow || mark.Target.Kind == HintTargetColumn ||
					mark.Target.Kind == HintTargetBox {
					t.Fatalf("chain scene paints broad unit scope: %+v", mark)
				}
			}
			switch key {
			case "simple-coloring":
				if len(plan.Steps[1].Links) == 0 || plan.Steps[1].Nodes[0].Group == "" {
					t.Fatalf("coloring plan does not teach alternating groups: %+v", plan.Steps)
				}
			case "x-cycles":
				if len(plan.Steps[1].Links) < 3 || plan.Steps[1].Nodes[0].Order == 0 {
					t.Fatalf("cycle plan does not teach alternating links: %+v", plan.Steps)
				}
			case "xy-chain":
				for _, step := range plan.Steps {
					if len(step.Narration) == 0 {
						t.Fatalf("XY-chain step lacks structured narration: %+v", step)
					}
				}
				if len(plan.Steps[1].Links) == 0 || plan.Steps[1].Links[0].Kind != solver.LinkImplication {
					t.Fatalf("XY-chain plan does not teach the endpoint implication: %+v", plan.Steps)
				}
			}
		})
	}
}

func TestBoxLineReductionUsesItsSpecificRelationshipName(t *testing.T) {
	source := solver.UnitRef{Kind: solver.UnitRow, Index: 3}
	box := solver.UnitRef{Kind: solver.UnitBox, Index: 3}
	move := &solver.Move{
		Cell:      core.NewCell(core.NewPosition(4, 2), 7),
		Technique: "pointing-pair",
		Evidence: &solver.Evidence{
			Unit:  &source,
			Units: []solver.UnitRef{source, box},
			Premises: []solver.CandidateGroup{
				{Position: core.NewPosition(3, 0), Values: []int{9}},
				{Position: core.NewPosition(3, 1), Values: []int{9}},
			},
		},
	}
	steps := intersectionSteps(move)
	if len(steps) != 3 || steps[0].Marks[0].Target.Kind != HintTargetRow ||
		steps[1].Marks[len(steps[1].Marks)-1].Target.Kind != HintTargetBox {
		t.Fatalf("unexpected box-line scene: %+v", steps)
	}
	if steps[0].Message.ID != "hint.intersection-confinement.observe" ||
		steps[1].Message.ID != "hint.intersection-confinement.explain" {
		t.Fatalf("box-line relationship uses generic descriptors: %+v", steps)
	}
}

func TestUnsupportedStrategicEvidenceReturnsNoPlan(t *testing.T) {
	move := &solver.Move{
		Cell:      core.NewCell(core.NewPosition(0, 0), 1),
		Technique: "simple-coloring",
		Evidence:  &solver.Evidence{},
	}
	if plan := composeHintPlan(Snapshot{}, nil, move); plan != nil {
		t.Fatalf("unsupported strategic evidence produced a plan: %+v", plan)
	}
}
