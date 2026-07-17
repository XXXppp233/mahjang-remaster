package mahjong

type Rule struct {
	Golden, Season, Flowers, Winds bool
	Chun, Hatsu, Haku              bool
	Chow, Pong, Kong, Expose       bool
	SevenPair                      bool
	Total                          int
}

var Fuzhou = Rule{
	Golden: true, Winds: true, Chun: true, Hatsu: true, Haku: true,
	Chow: true, Pong: true, Kong: true, Expose: true, Total: 16,
}

var SortOrder = map[string]int{
	"Golden": 0,
	"Man1":   1, "Man2": 2, "Man3": 3, "Man4": 4, "Man5": 5, "Man6": 6, "Man7": 7, "Man8": 8, "Man9": 9,
	"Pin1": 11, "Pin2": 12, "Pin3": 13, "Pin4": 14, "Pin5": 15, "Pin6": 16, "Pin7": 17, "Pin8": 18, "Pin9": 19,
	"Sou1": 21, "Sou2": 22, "Sou3": 23, "Sou4": 24, "Sou5": 25, "Sou6": 26, "Sou7": 27, "Sou8": 28, "Sou9": 29,
	"Ton": 31, "Nan": 32, "Shaa": 33, "Pei": 34, "Chun": 35, "Hatsu": 36, "Haku": 37,
	"spring": 38, "summer": 39, "autumn": 40, "winter": 41,
	"plum": 42, "orchid": 43, "bamboo": 44, "chrysanthemum": 45,
}
