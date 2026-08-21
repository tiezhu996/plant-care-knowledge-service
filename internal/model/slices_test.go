package model_test

import (
	"testing"

	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

func TestTakeDoesNotAlias(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := util.Take(src, 3)
	if len(got) != 3 {
		t.Fatalf("util.Take(10,3) len=%d", len(got))
	}
	got[0] = 99
	if src[0] != 1 {
		t.Fatalf("Take must copy; mutating the result corrupted src[0]=%d", src[0])
	}
}
