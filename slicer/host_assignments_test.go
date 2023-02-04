package slicer

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHostMapping(t *testing.T) {
	tests := []struct {
		name        string
		assignments map[string][]Range
		wantErr     bool
	}{
		{"nil", nil, true},
		{"single assignment", map[string][]Range{"name": {{0, math.MaxInt64}}}, false},
		{"valid disjoint chunks", map[string][]Range{
			"first":  {{0, 99}, {200, 299}},
			"second": {{100, 199}, {300, math.MaxInt64}},
		}, false},
		{"invalid Range", map[string][]Range{"backwards": {{2, 1}}}, true},
		{"simple incomplete Range", map[string][]Range{"too small": {{1, 2}}}, true},
		{"disjoint incomplete Range", map[string][]Range{
			"first":  {{0, 99}, {200, 201}},
			"second": {{100, 199}, {300, math.MaxInt64}},
		}, true},
		{"disjoint incomplete Range 2", map[string][]Range{
			"first":  {{0, 99}, {201, 299}},
			"second": {{100, 199}, {300, math.MaxInt64}},
		}, true},
		{"overlapping range", map[string][]Range{
			"first":  {{0, 99}, {120, 299}},
			"second": {{100, 199}, {300, math.MaxInt64}},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHostAssignments(tt.assignments)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHostMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_hostMapping_GetKeyHost(t *testing.T) {
	type query struct {
		searchKey int64
		wantHost  string
	}
	tests := []struct {
		name        string
		assignments map[string][]Range
		queries     []query
	}{
		{"single assignment", map[string][]Range{"server": {{0, math.MaxInt64}}}, []query{{1, "server"}}},
		{"disjoint chunks", map[string][]Range{
			"first":  {{0, 99}, {200, 299}},
			"second": {{100, 199}, {300, math.MaxInt64}},
		}, []query{{22, "first"}, {150, "second"}, {230, "first"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm, err := NewHostAssignments(tt.assignments)
			if err != nil {
				t.Errorf("NewHostMapping() error = %v", err)
				return
			}
			for _, query := range tt.queries {
				assert.Equal(t, query.wantHost, hm.getKeyHost(query.searchKey).Host)
			}
		})
	}
}

func TestHash(t *testing.T) {
	// Just some precomputed hashes to ensure stability
	assert.Equal(t, int64(4316084372001321715), Hash("steady"))
	assert.Equal(t, int64(4539478222691259463), Hash("fixed"))
	assert.Equal(t, int64(1940846785606929000), Hash("stable"))
	assert.Equal(t, int64(1995353473133317714), Hash("unmoving"))
	assert.Equal(t, int64(6062514253447166957), Hash("slicer"))
	assert.Equal(t, int64(8656414110124581126), Hash("$%^&*("))
}

func TestSerialize(t *testing.T) {
	mapping, err := NewHostAssignments(map[string][]Range{
		"first":  {{0, 99}, {200, 299}},
		"second": {{100, 199}, {300, math.MaxInt64}}})
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := mapping.Serialize(&buf); err != nil {
		t.Fatal(err)
	}

	new, err := NewFromSerialized(&buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, new.assignments, mapping.assignments)

	{
		_, err := NewFromSerialized(&bytes.Buffer{})
		assert.Error(t, err)
	}
}
