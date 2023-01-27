package slicer

import (
	"os"
	"testing"
)

func TestSlicer_AddLocation(t *testing.T) {
	slicer := &Assigner{}

	if err := slicer.AddHost("one", nil); err != nil {
		t.Fatal(err)
	}

	if err := slicer.AddHost("two", nil); err != nil {
		t.Fatal(err)
	}

	slicer.hostAssignments.Serialize(os.Stdout)
}
