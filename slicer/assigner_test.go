package slicer

import (
	"os"
	"testing"
)

func TestSlicer_AddLocation(t *testing.T) {
	slicer := &Assigner{}

	if err := slicer.AddLocation("one", nil); err != nil {
		t.Fatal(err)
	}

	if err := slicer.AddLocation("two", nil); err != nil {
		t.Fatal(err)
	}

	slicer.hostAssignments.Serialize(os.Stdout)
}
