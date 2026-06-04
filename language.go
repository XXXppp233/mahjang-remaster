package main

type Language struct {
	ErrorCode map[int]string
}

var Chinese = Language{
	ErrorCode: map[int]string{
		0:  "成功",
		10: "游戏已经开始，无法修改规则",
		11: "游戏已经开始，无法添加玩家",
		12: "房间已满，无法添加玩家",
		13: "游戏已经开始，无法移除玩家",
		21: "游戏未开始，无法吃碰杠",
		22: "无法吃碰自己打出的牌",
		23: "无法吃碰金牌",
		24:  "不同牌面无法互吃",
		999: "未知错误",
	},
}
