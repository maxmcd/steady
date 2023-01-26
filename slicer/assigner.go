package slicer

import (
	"math"
	"sort"
)

type Assigner struct {
	hostAssignments *HostAssignments
}

func (s *Assigner) Host(key string) string {
	return s.hostAssignments.GetHost(key).Host
}

func (s *Assigner) Assignments() map[string][]Range {
	return s.hostAssignments.Assignments()
}

func (s *Assigner) AddLocation(name string, liveKeys []string) (err error) {
	if s.hostAssignments == nil {
		s.hostAssignments, err = NewHostAssignments(map[string][]Range{
			name: {{0, math.MaxInt64}},
		})
		return err
	}

	// Just consistent hashing for now...

	chunks := [][]Range{}
	newCount := int64(len(s.hostAssignments.assignments) + 1)
	size := math.MaxInt64 / newCount
	for i := int64(0); i < newCount; i++ {
		chunks = append(chunks, []Range{{size * i, size*i + size - 1}})
	}
	chunks[len(chunks)-1][0].End = math.MaxInt64

	hostNames := []string{name}
	for n := range s.hostAssignments.assignments {
		hostNames = append(hostNames, n)
	}
	sort.Strings(hostNames)

	assignments := map[string][]Range{}
	for i, n := range hostNames {
		assignments[n] = chunks[i]
	}

	return s.hostAssignments.NewAssignments(assignments)
}
