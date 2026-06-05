package main

// var tiles = map[string]int{
// "Man1": 1,
// "Man2": 2,
// "Man3": 3,
// "Man4": 4,
// "Man5": 5,
// "Man6": 6,
// "Man7": 7,
// "Man8": 8,
// "Man9": 9,
// "Pin1": 1,
// }

type tile struct {
	Face string
	Num  int
}

func (t *tile) Parse(s string) tile {
	if len(s) >= 4 && (s[0:3] == "Man" || s[0:3] == "Pin" || s[0:3] == "Sou") {
		num := int(s[3] - '0')
		return tile{Face: s[0:3], Num: num}
	}
	return tile{Face: s, Num: 0}
}

type Player struct {
	Cookie          string
	Uuid            string
	Name            string
	IP              string
	Grade           Grade
	Room            *Room
	Ready           bool
	CharactersGroup string
	Character       string
	Playing         bool // 当玩家掉线时，如果 Playing 为 True 则会等待玩家重连回到房间，
	Online          bool // 玩家是否在线，掉线时会设置为 false，重连时设置为 true
}
type Grade struct {
	First  int
	Second int
	Third  int
	Fourth int
}

type Rule struct {
	Golden    bool // 是否启用金牌/癞子
	Season    bool // 是否启用春夏秋冬牌
	Flowers   bool // 是否启用梅兰竹菊牌
	Winds     bool // 是否启用东南西北牌
	Chun      bool // 是否启用中
	Hatsu     bool // 是否启用发财
	Haku      bool // 是否启用白板
	Chow      bool // 是否允许吃
	Pong      bool // 是否允许碰
	Kong      bool // 是否允许杠
	Expose    bool // True 为不允许暗杠，即杠自己的牌也会显示牌面
	SevenPair bool // 七对子，仅在 Total == 14 时生效
	Total     int  // 玩家手牌数量
}

type OfflineList map[string]*Room // [uuid]: Room
