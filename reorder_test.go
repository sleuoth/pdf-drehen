package main

import (
	"reflect"
	"testing"
)

func TestBuildOrder_Even40(t *testing.T) {
	got := buildOrderIndices(40)
	// expect: 0,20,1,21,2,22,...,19,39
	want := make([]int, 0, 40)
	for i := 0; i < 20; i++ {
		want = append(want, i, 20+i)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order mismatch for 40 pages.\n got=%v\nwant=%v", got[:10], want[:10])
	}
	// Spot-check a few positions
	if got[0] != 0 || got[1] != 20 || got[2] != 1 || got[3] != 21 {
		t.Fatalf("unexpected head: %v", got[:6])
	}
	if got[len(got)-2] != 19 || got[len(got)-1] != 39 {
		t.Fatalf("unexpected tail: %v", got[len(got)-6:])
	}
}

func TestBuildOrder_Odd5(t *testing.T) {
	// With 5 pages we conceptually pad to 6: expect 0,3,1,4,2,5
	got := buildOrderIndices(5)
	want := []int{0, 3, 1, 4, 2, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order mismatch for 5 pages (with pad).\n got=%v\nwant=%v", got, want)
	}
	// Length should be even
	if len(got)%2 != 0 {
		t.Fatalf("expected even length, got %d", len(got))
	}
}

func TestBuildOrder_Minimal2(t *testing.T) {
	got := buildOrderIndices(2)
	want := []int{0, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order mismatch for 2 pages.\n got=%v\nwant=%v", got, want)
	}
}

func TestBuildOrder_ZeroAndOne(t *testing.T) {
	// Edge cases: 0 -> empty; 1 -> pad to 2 => 0,1
	if got := buildOrderIndices(0); len(got) != 0 {
		t.Fatalf("expected empty for 0, got %v", got)
	}
	if got := buildOrderIndices(1); !reflect.DeepEqual(got, []int{0, 1}) {
		t.Fatalf("expected [0 1] for 1 (pad), got %v", got)
	}
}
