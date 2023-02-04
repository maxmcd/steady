package slicer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlicer_AddLocation(t *testing.T) {
	slicer := &Assigner{}

	if err := slicer.AddHost("one", nil); err != nil {
		t.Fatal(err)
	}

	if err := slicer.AddHost("two", nil); err != nil {
		t.Fatal(err)
	}

	if err := slicer.hostAssignments.Serialize(os.Stdout); err != nil {
		t.Fatal(err)
	}
}

func TestSlicerDistinct(t *testing.T) {
	// Ensure we have a key that is known to be migrated to the second server
	// once it has been added
	slicer := &Assigner{}

	if err := slicer.AddHost("one", nil); err != nil {
		t.Fatal(err)
	}
	first := slicer.Host("whee")

	if err := slicer.AddHost("two", nil); err != nil {
		t.Fatal(err)
	}
	second := slicer.Host("whee")

	assert.NotEqual(t, first, second)
}
