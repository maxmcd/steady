package slicer

import (
	"math"
)

// Assigner maintains hosts and their assignments.
type Assigner struct {
	hostAssignments *HostAssignments

	// keep a consistent order so that it's easier to write deterministic tests,
	// remove when we have more complicated slicing logic and can write tests
	// around explicit calculated migrations
	hostnames []string
}

// Host provides a host's identifier for a given key.
func (s *Assigner) Host(key string) string {
	return s.hostAssignments.GetHost(key)
}

func (s *Assigner) Assignments() map[string][]Range {
	return s.hostAssignments.Assignments()
}

func (s *Assigner) AddHost(name string, liveKeys []string) (err error) {
	if s.hostAssignments == nil {
		s.hostnames = []string{name}
		s.hostAssignments, err = NewHostAssignments(map[string][]Range{
			name: {{0, math.MaxInt64}},
		})
		return err
	}

	// TODO: data race?
	s.hostnames = append(s.hostnames, name)

	// Just consistent hashing for now...
	chunks := [][]Range{}
	newCount := int64(len(s.hostAssignments.assignments) + 1)
	size := math.MaxInt64 / newCount
	for i := int64(0); i < newCount; i++ {
		chunks = append(chunks, []Range{{size * i, size*i + size - 1}})
	}
	chunks[len(chunks)-1][0].End = math.MaxInt64

	assignments := map[string][]Range{}
	for i, n := range s.hostnames {
		assignments[n] = chunks[i]
	}

	return s.hostAssignments.NewAssignments(assignments)
}
