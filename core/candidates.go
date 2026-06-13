package core

import "math/bits"

// CandidateSet is a bitfield representing candidate values 1–9 for a cell.
// Bit i (1 ≤ i ≤ 9) is set when value i is a candidate.
type CandidateSet uint16

// allCandidates has bits 1–9 set (0x3FE = 0b0000_0011_1111_1110).
const allCandidates CandidateSet = 0x3FE

// Has reports whether value v is a candidate.
func (c CandidateSet) Has(v int) bool {
	return c&(1<<v) != 0
}

// Add adds value v to the candidate set.
func (c *CandidateSet) Add(v int) {
	*c |= 1 << v
}

// Remove removes value v from the candidate set.
func (c *CandidateSet) Remove(v int) {
	*c &^= 1 << v
}

// Count returns the number of candidates in the set.
func (c CandidateSet) Count() int {
	return bits.OnesCount16(uint16(c))
}

// Single returns the sole candidate value and true if exactly one candidate
// remains, or (0, false) otherwise.
func (c CandidateSet) Single() (int, bool) {
	if bits.OnesCount16(uint16(c)) != 1 {
		return 0, false
	}
	return bits.TrailingZeros16(uint16(c)), true
}

// Values returns a slice of all candidate values in the set.
func (c CandidateSet) Values() []int {
	vals := make([]int, 0, c.Count())
	for v := 1; v <= 9; v++ {
		if c.Has(v) {
			vals = append(vals, v)
		}
	}
	return vals
}

// IsEmpty reports whether the candidate set contains no candidates.
func (c CandidateSet) IsEmpty() bool {
	return c == 0
}
