package mahjong

import "testing"

func TestCanHuWithGolden(t *testing.T) {
	tiles := []string{"Man1", "Man1", "Man1", "Man2", "Man3", "Man4", "Pin2", "Pin3", "Pin4", "Sou7", "Sou8", "Sou9", "Ton", "Ton", "Chun", "Chun"}
	if !CanHu(tiles, 1) {
		t.Fatal("expected wildcard hand to win")
	}
}
