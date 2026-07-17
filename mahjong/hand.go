package mahjong

import "fmt"

// CanHu evaluates the standard Fuzhou hand shape with Golden wildcards.
func CanHu(tiles []string, goldenCount int) bool {
	if goldenCount >= 3 {
		return true
	}
	for i := range tiles {
		if i+1 < len(tiles) && tiles[i] == tiles[i+1] {
			remaining := append([]string{}, tiles[:i]...)
			remaining = append(remaining, tiles[i+2:]...)
			if allSets(remaining, goldenCount) {
				return true
			}
		}
		if goldenCount >= 1 {
			remaining := append([]string{}, tiles[:i]...)
			remaining = append(remaining, tiles[i+1:]...)
			if allSets(remaining, goldenCount-1) {
				return true
			}
		}
	}
	return goldenCount == 2 && allSets(tiles, 0)
}

func allSets(tiles []string, goldenCount int) bool {
	if len(tiles) == 0 {
		return goldenCount%3 == 0
	}
	first := tiles[0]
	count := 0
	for _, value := range tiles {
		if value == first {
			count++
		}
	}
	if count >= 3 && allSets(tiles[3:], goldenCount) {
		return true
	}
	if count >= 2 && goldenCount >= 1 && allSets(tiles[2:], goldenCount-1) {
		return true
	}
	if goldenCount >= 2 && allSets(tiles[1:], goldenCount-2) {
		return true
	}

	face, number, ok := suitedTile(first)
	if ok && number <= 7 {
		second, third := fmt.Sprintf("%s%d", face, number+1), fmt.Sprintf("%s%d", face, number+2)
		i2, i3 := find(tiles, second), find(tiles, third)
		if i2 >= 0 && i3 >= 0 && allSets(remove(tiles, 0, i2, i3), goldenCount) {
			return true
		}
		if i2 >= 0 && goldenCount >= 1 && allSets(remove(tiles, 0, i2), goldenCount-1) {
			return true
		}
		if i3 >= 0 && goldenCount >= 1 && allSets(remove(tiles, 0, i3), goldenCount-1) {
			return true
		}
	}
	return false
}

func suitedTile(value string) (string, int, bool) {
	if len(value) != 4 {
		return "", 0, false
	}
	face := value[:3]
	if face != "Man" && face != "Pin" && face != "Sou" {
		return "", 0, false
	}
	if value[3] < '1' || value[3] > '9' {
		return "", 0, false
	}
	return face, int(value[3] - '0'), true
}

func find(values []string, wanted string) int {
	for i, value := range values {
		if value == wanted {
			return i
		}
	}
	return -1
}

func remove(values []string, indices ...int) []string {
	removed := make(map[int]bool, len(indices))
	for _, i := range indices {
		removed[i] = true
	}
	result := make([]string, 0, len(values)-len(indices))
	for i, value := range values {
		if !removed[i] {
			result = append(result, value)
		}
	}
	return result
}
