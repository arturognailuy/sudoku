package solver

var strategyTierNames = []string{"easy", "medium", "hard", "expert", "evil"}

var strategyTierTechniques = map[string][]string{
	"easy":   {"naked-single", "hidden-single"},
	"medium": {"naked-pair", "naked-triple", "pointing-pair", "hidden-pair"},
	"hard":   {"x-wing", "xy-wing", "hidden-triple", "w-wing"},
	"expert": {"swordfish", "naked-quad", "simple-coloring", "hidden-quad", "xyz-wing"},
	"evil":   {"jellyfish", "bug-plus-one", "unique-rectangle", "unique-rectangle-2", "unique-rectangle-3", "unique-rectangle-4", "x-cycles", "xy-chain"},
}

var strategyTechniqueTiers = func() map[string]int {
	result := make(map[string]int)
	for tier, name := range strategyTierNames {
		for _, technique := range strategyTierTechniques[name] {
			result[technique] = tier
		}
	}
	return result
}()

// StrategyTierNames returns the canonical difficulty hierarchy from lowest to
// highest. The returned slice is detached from package state.
func StrategyTierNames() []string {
	return append([]string(nil), strategyTierNames...)
}

// StrategySolverKeysForTier returns the canonical stable solver order for a
// difficulty tier. The returned slice is detached from package state.
func StrategySolverKeysForTier(name string) []string {
	return append([]string(nil), strategyTierTechniques[name]...)
}

// StrategyTierForTechnique returns a technique's tier name and hierarchy index.
func StrategyTierForTechnique(technique string) (string, int, bool) {
	tier, ok := strategyTechniqueTiers[technique]
	if !ok {
		return "", -1, false
	}
	return strategyTierNames[tier], tier, true
}
