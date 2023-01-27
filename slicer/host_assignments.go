package slicer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"sort"
	"sync"
)

type Range struct {
	Start int64
	End   int64
}

type RangeAndHost struct {
	Host  string
	Range Range
}

// HostAssignments stores hosts and their corresponding ranges of hash values.
// Each host gets a set of key ranges that it owns. This struct handles
// building, validating and querying those slice ranges.
type HostAssignments struct {
	lock        sync.RWMutex
	assignments map[string][]Range
	ranges      []RangeAndHost
}

// NewAssignments takes a new set of assignments. If the assignment ranges are
// not complete, or are overlapping, this function will return an error.
func (hm *HostAssignments) NewAssignments(assignments map[string][]Range) error {
	var ranges []RangeAndHost
	for host, hostRanges := range assignments {
		for _, r := range hostRanges {
			if r.End < r.Start {
				return fmt.Errorf("Range %v is invalid", r)
			}
			ranges = append(ranges, RangeAndHost{Host: host, Range: r})
		}
	}
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Range.Start < ranges[j].Range.Start
	})

	var last Range
	var expectedNextStart int64 = 0
	for _, r := range ranges {
		if r.Range.Start != expectedNextStart {
			return fmt.Errorf(
				"Assignment range is invalid. Expected next range in series to begin with %d, but it began with %d",
				expectedNextStart, r.Range.Start)
		}
		expectedNextStart = r.Range.End + 1
		last = r.Range
	}
	if last.End != math.MaxInt64 {
		return fmt.Errorf(
			"Assignment range is invalid. Expected range to end with %d, but it ended with %d",
			math.MaxInt64, last.End)
	}

	// Don't assign and edit values until all validation is complete
	hm.lock.Lock()
	hm.ranges = ranges
	hm.assignments = assignments
	hm.lock.Unlock()
	return nil
}

// GetHost get a host's identifier for a given key
func (hm *HostAssignments) GetHost(key string) string {
	return hm.getKeyHost(Hash(key)).Host
}

// Assignments returns a full copy of the internal assignments
func (hm *HostAssignments) Assignments() map[string][]Range {
	assignments := map[string][]Range{}
	hm.lock.RLock()
	for k, v := range hm.assignments {
		assignments[k] = v
	}
	hm.lock.RUnlock()
	return assignments
}

func (hm *HostAssignments) getKeyHost(hash int64) RangeAndHost {
	hm.lock.RLock()
	defer hm.lock.RUnlock()
	low := 0
	high := len(hm.ranges) - 1

	for {
		median := (low + high) / 2
		rng := hm.ranges[median]
		if hash >= rng.Range.Start && hash <= rng.Range.End {
			return rng
		}
		if hash > rng.Range.Start {
			low = median + 1
		} else {
			high = median - 1
		}
	}
}

func (hm *HostAssignments) Serialize(w io.Writer) error {
	hm.lock.RLock()
	defer hm.lock.RUnlock()
	return json.NewEncoder(w).Encode(hm.assignments)
}

// NewHostAssignments will create a new HostAssignments from a complete set of
// assignments. If the assignment ranges are not complete, or are overlapping,
// this function will return an error.
func NewHostAssignments(assignments map[string][]Range) (*HostAssignments, error) {
	hm := &HostAssignments{}
	return hm, hm.NewAssignments(assignments)
}

func NewFromSerialized(r io.Reader) (*HostAssignments, error) {
	var assignments map[string][]Range
	if err := json.NewDecoder(r).Decode(&assignments); err != nil {
		return nil, err
	}
	return NewHostAssignments(assignments)
}

func Hash(key string) int64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(key))

	i := int64(binary.BigEndian.Uint64(h.Sum(nil)))
	if i < 0 {
		return i * -1
	}
	return i
}
