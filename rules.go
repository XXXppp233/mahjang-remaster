package main

import mahjonglib "mahjong-server/mahjong"

// 胡牌优先级，例如 0 号玩家打出牌时，胡牌优先级最高的是1号玩家, 2,3 号玩家会被劫胡
var HuPri = map[int][]int{
	0: []int{3, 0, 1, 2},
	1: []int{2, 3, 0, 1},
	2: []int{1, 2, 3, 0},
	3: []int{0, 1, 2, 3},
}

var RuleList = map[int]Rule{
	0: Fuzhou,
	1: Japanese,
}

var RuleNames = map[int]string{
	0: "福州麻将",
	1: "日本麻将",
}

var Fuzhou = mahjonglib.Fuzhou

var Japanese = Rule{
	Golden:    false,
	Season:    false,
	Flowers:   false,
	Winds:     true,
	Chun:      true,
	Hatsu:     true,
	Haku:      true,
	Chow:      true,
	Pong:      true,
	Kong:      true,
	Expose:    true,
	SevenPair: true,
	Total:     13,
}

var SortRule = mahjonglib.SortOrder
