package mahjong

import (
	"reflect"
	"testing"
)

func TestShuffleIsDeterministic(t *testing.T) {
	a := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := append([]int(nil), a...)
	Shuffle(a, 666)
	Shuffle(b, 666)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed produced different results")
	}
	want := []int{4, 3, 9, 2, 7, 6, 0, 8, 1, 5}
	if !reflect.DeepEqual(a, want) {
		t.Fatalf("protocol vector changed: got %v", a)
	}
}
