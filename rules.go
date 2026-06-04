package main

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

var Fuzhou = Rule{
	Golden:    true,
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
	SevenPair: false,
	Total:     16,
}

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

var SortRule = map[string]int{
	"Man1":          1,
	"Man2":          2,
	"Man3":          3,
	"Man4":          4,
	"Man5":          5,
	"Man6":          6,
	"Man7":          7,
	"Man8":          8,
	"Man9":          9,
	"Pin1":          11,
	"Pin2":          12,
	"Pin3":          13,
	"Pin4":          14,
	"Pin5":          15,
	"Pin6":          16,
	"Pin7":          17,
	"Pin8":          18,
	"Pin9":          19,
	"Sou1":          21,
	"Sou2":          22,
	"Sou3":          23,
	"Sou4":          24,
	"Sou5":          25,
	"Sou6":          26,
	"Sou7":          27,
	"Sou8":          28,
	"Sou9":          29,
	"Ton":           31, // 东
	"Nan":           32, // 南
	"Shaa":          33, // 西
	"Pei":           34, // 北
	"Chun":          35, // 中
	"Hatsu":         36, // 发
	"Haku":          37, // 白
	"spring":        38,
	"summer":        39,
	"autumn":        40,
	"winter":        41,
	"plum":          42,
	"orchid":        43,
	"bamboo":        44,
	"chrysanthemum": 45,
	"Golden":        0,
}
