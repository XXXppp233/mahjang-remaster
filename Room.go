package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Room struct {
	Name      string // userinput
	Id        string
	Player    []*Player // 0 号位是房主，其他人可以自由交换位置，房主交换位置会转移房主的资格
	Watcher   []*Player
	Playing   bool // 游戏是否正在进行
	Starting  bool
	RoomRule  RoomRule
	GameState GameState
}
type RoomRule struct {
	Rule         int  // 0 代表标准福州麻将
	ShowCritical bool // 是否在 吃碰杠等操作后播放动画
	SkipOffline  bool // 跳过掉线玩家的出牌阶段，默认打出新牌，如果为 false，则会正常等待 10 秒钟以给玩家重连的机会，LoginRequire 为 False 时忽略规则更新强制为 True
	MaxWaiting   int  // 等待玩家出牌的时间，5~20，超过 20 则代表无限等待时间
}
type GameState struct {
	Wall            []string // 牌墙
	GoldenTile      string
	PlayerInfo      []PlayerInfo
	CurrentUser     int    // 当前未出牌的玩家 0 1 2 3
	WaitingResponse bool   // 玩家出牌后会等待两秒其他玩家操作，即使没有玩家能够操作
	Out             string // 用户打出的牌
	ActionQueue     ActionQueue
}
type PlayerInfo struct {
	Uuid            string
	Name            string
	Online          bool
	Index           int // 0 1 2 3
	CharactersGroup string
	Character       string
	Ready           bool
	Lock            []string // 吃碰杠后的牌，无法再操作，不会用来计算胡牌
	Hands           []string // 手牌
	New             string   // 新摸到的牌
	Discarded       []string
	Score           int
}

type ActionQueue []Action
type RoomList map[string]*Room // roomid: Room

type Action struct {
	Type   int // 1 chow, 2 pong, 3 kong, 4 hu
	Pri    int // 优先级，1~3 胡，5杠，6碰，7 吃
	Player int // Playerinfo.Index
	Selec  []int
}

func isGoldenTile(tile string) bool {
	return tile == "Golden"
}

func (r *Room) Update(rule RoomRule) (bool, int) {
	if r.Playing {
		return false, 10
	}
	if _, ok := RuleList[rule.Rule]; !ok {
		return false, 14
	}
	if !GlobalConf.LoginRequire {
		rule.SkipOffline = true
	}
	r.RoomRule = rule
	return true, 0
}

func (r *Room) AddUser(user *Player) (bool, int) {
	if r.Playing {
		return false, 11
	}
	if len(r.Player) == 4 {
		return false, 12
	}
	for _, p := range r.Player {
		if p.Uuid == user.Uuid {
			return true, 0
		}
	}
	r.Player = append(r.Player, user)
	return true, 0
}

func (r *Room) RemoveUser(user *Player) (bool, int) {
	if r.Playing {
		for _, p := range r.Player {
			if p.Uuid == user.Uuid {
				p.Online = false
			}
		}
		return false, 13
	}
	for i, p := range r.Player {
		if p.Uuid == user.Uuid {
			r.Player = append(r.Player[:i], r.Player[i+1:]...)
			user.Room = nil
			return true, 0
		}
	}
	return true, 0
}

// 无法使用金牌来吃碰杠，胡牌计算需要额外的检查函数
func (r *Room) CheckChow(p *PlayerInfo, selec [2]int) (bool, int) {
	if !r.Playing {
		return false, 21
	}
	if ok, code := r.canRespondToOut(p); !ok {
		return false, code
	}
	if r.hasQueuedAction(p.Index) {
		return false, 24
	}
	if !validIndices(len(p.Hands), selec[0], selec[1]) {
		return false, 20
	}
	// 只能吃上家打出的牌
	if (r.GameState.CurrentUser+1)%4 != p.Index {
		return false, 22
	}
	if isGoldenTile(p.Hands[selec[0]]) || isGoldenTile(p.Hands[selec[1]]) {
		return false, 23
	}
	res, errCode := r.CanChow([2]string{p.Hands[selec[0]], p.Hands[selec[1]]})
	if !res {
		return false, errCode
	}
	r.GameState.ActionQueue = append(r.GameState.ActionQueue, Action{
		Type:   1,
		Pri:    7,
		Player: p.Index,
		Selec:  []int{selec[0], selec[1]},
	})

	return true, 0
}

func (r *Room) CanChow(tiles [2]string) (bool, int) {
	var t tile
	out, tile1, tile2 := t.Parse(r.GameState.Out), t.Parse(tiles[0]), t.Parse(tiles[1])
	if out.Num == 0 || tile1.Num == 0 || tile2.Num == 0 {
		return false, 24
	}
	if tile1.Face != out.Face || tile2.Face != out.Face {
		return false, 25
	}

	nums := []int{out.Num, tile1.Num, tile2.Num}
	sort.Ints(nums)

	if nums[0]+1 == nums[1] && nums[1]+1 == nums[2] {
		return true, 0
	}
	return false, 26
}

func (r *Room) CheckPong(p *PlayerInfo, selec [2]int) (bool, int) {
	if !r.Playing {
		return false, 21
	}
	if ok, code := r.canRespondToOut(p); !ok {
		return false, code
	}
	if r.hasQueuedAction(p.Index) {
		return false, 24
	}
	if !validIndices(len(p.Hands), selec[0], selec[1]) {
		return false, 20
	}
	if isGoldenTile(p.Hands[selec[0]]) || isGoldenTile(p.Hands[selec[1]]) {
		return false, 23
	}
	if p.Hands[selec[0]] != r.GameState.Out || p.Hands[selec[1]] != r.GameState.Out {
		return false, 31
	}
	r.GameState.ActionQueue = append(r.GameState.ActionQueue, Action{
		Type:   2,
		Pri:    6,
		Player: p.Index,
		Selec:  []int{selec[0], selec[1]},
	})
	return true, 0
}

func (r *Room) CheckKong(p *PlayerInfo, selec []int) (bool, int) {
	if !r.Playing {
		return false, 21
	}
	if r.GameState.CurrentUser != p.Index {
		if ok, code := r.canRespondToOut(p); !ok {
			return false, code
		}
	}
	if r.GameState.CurrentUser != p.Index && r.hasQueuedAction(p.Index) {
		return false, 24
	}
	if !validIndices(len(p.Hands), selec...) {
		return false, 20
	}
	// 暗杠/补杠 (self turn)
	if r.GameState.CurrentUser == p.Index {
		if len(selec) == 4 { // 暗杠
			tile := p.Hands[selec[0]]
			for _, i := range selec {
				if p.Hands[i] != tile {
					return false, 41
				}
			}
			r.ExecuteAction(Action{
				Type:   3,
				Pri:    5,
				Player: p.Index,
				Selec:  selec,
			})
			return true, 0
		}
	} else { // 明杠 (other's turn)
		if len(selec) != 3 {
			return false, 42
		}
		for _, i := range selec {
			if p.Hands[i] != r.GameState.Out {
				return false, 43
			}
		}
		r.GameState.ActionQueue = append(r.GameState.ActionQueue, Action{
			Type:   3,
			Pri:    5,
			Player: p.Index,
			Selec:  selec,
		})
	}
	return true, 0
}

func (r *Room) CheckHu(p *PlayerInfo) (bool, int) {
	if !r.Playing {
		return false, 21
	}
	if r.GameState.CurrentUser != p.Index {
		if ok, code := r.canRespondToOut(p); !ok {
			return false, code
		}
	}
	if r.GameState.CurrentUser != p.Index && r.hasQueuedAction(p.Index) {
		return false, 24
	}

	allTiles := append([]string{}, p.Hands...)
	if r.GameState.CurrentUser == p.Index {
		if p.New != "" {
			allTiles = append(allTiles, p.New)
		}
	} else {
		allTiles = append(allTiles, r.GameState.Out)
	}

	goldenCount := 0
	normalTiles := []string{}
	for _, t := range allTiles {
		if isGoldenTile(t) {
			goldenCount++
		} else {
			normalTiles = append(normalTiles, t)
		}
	}

	sort.Slice(normalTiles, func(i, j int) bool {
		return SortRule[normalTiles[i]] < SortRule[normalTiles[j]]
	})

	if r.CanHu(normalTiles, goldenCount) {
		action := Action{
			Type:   4,
			Pri:    r.huPriority(p.Index),
			Player: p.Index,
		}
		if r.GameState.CurrentUser == p.Index {
			r.ExecuteAction(action)
		} else {
			r.GameState.ActionQueue = append(r.GameState.ActionQueue, action)
		}
		return true, 0
	}

	return false, 51
}

func (r *Room) CanHu(tiles []string, goldenCount int) bool {
	// 三金
	if goldenCount >= 3 {
		return true
	}

	// 尝试每个可能的雀头
	for i := 0; i < len(tiles); i++ {
		// 1. 两张相同的普通牌
		if i+1 < len(tiles) && tiles[i] == tiles[i+1] {
			remaining := append([]string{}, tiles[:i]...)
			remaining = append(remaining, tiles[i+2:]...)
			if r.isAllSets(remaining, goldenCount) {
				return true
			}
		}
		// 2. 一张普通牌 + 一张金牌
		if goldenCount >= 1 {
			remaining := append([]string{}, tiles[:i]...)
			remaining = append(remaining, tiles[i+1:]...)
			if r.isAllSets(remaining, goldenCount-1) {
				return true
			}
		}
	}
	// 3. 两张金牌
	if goldenCount == 2 {
		if r.isAllSets(tiles, goldenCount-2) {
			return true
		}
	}
	return false
}

func (r *Room) isAllSets(tiles []string, goldenCount int) bool {
	if len(tiles) == 0 {
		return goldenCount%3 == 0
	}

	first := tiles[0]
	// 尝试刻子
	count := 0
	for _, t := range tiles {
		if t == first {
			count++
		}
	}
	// 1. 三张相同
	if count >= 3 {
		if r.isAllSets(tiles[3:], goldenCount) {
			return true
		}
	}
	// 2. 两张相同 + 一张金牌
	if count >= 2 && goldenCount >= 1 {
		if r.isAllSets(tiles[2:], goldenCount-1) {
			return true
		}
	}
	// 3. 一张相同 + 两张金牌
	if count >= 1 && goldenCount >= 2 {
		if r.isAllSets(tiles[1:], goldenCount-2) {
			return true
		}
	}

	// 尝试顺子
	var tObj tile
	ft := tObj.Parse(first)
	if ft.Num != 0 && ft.Num <= 7 {
		t2 := fmt.Sprintf("%s%d", ft.Face, ft.Num+1)
		t3 := fmt.Sprintf("%s%d", ft.Face, ft.Num+2)

		idx2, idx3 := -1, -1
		for i, t := range tiles {
			if t == t2 && idx2 == -1 {
				idx2 = i
			} else if t == t3 && idx3 == -1 {
				idx3 = i
			}
		}

		if idx2 != -1 && idx3 != -1 {
			rem := removeIndices(tiles, 0, idx2, idx3)
			if r.isAllSets(rem, goldenCount) {
				return true
			}
		}
		if idx2 != -1 && goldenCount >= 1 {
			rem := removeIndices(tiles, 0, idx2)
			if r.isAllSets(rem, goldenCount-1) {
				return true
			}
		}
		if idx3 != -1 && goldenCount >= 1 {
			rem := removeIndices(tiles, 0, idx3)
			if r.isAllSets(rem, goldenCount-1) {
				return true
			}
		}
	}

	// 如果金牌足够直接补一个顺子（以当前第一张牌ft配合金牌）
	// 情况已经在上面 idx2/idx3 为 -1 且 goldenCount 足够时处理了一部分
	// 还可以有 ft + gold + gold
	if goldenCount >= 2 {
		if r.isAllSets(tiles[1:], goldenCount-2) {
			return true
		}
	}

	return false
}

func (r *Room) Broadcast() {
	targets := make([]*Player, 0, len(r.Player)+len(r.Watcher))
	targets = append(targets, r.Player...)
	targets = append(targets, r.Watcher...)

	for _, p := range targets {
		state, _ := json.Marshal(r.GameStateFor(p.Uuid))
		sendSSETo(p.Uuid, string(state))
	}
}

func (r *Room) BroadcastEvent(payload any) {
	state, _ := json.Marshal(payload)

	targets := make([]*Player, 0, len(r.Player)+len(r.Watcher))
	targets = append(targets, r.Player...)
	targets = append(targets, r.Watcher...)

	for _, p := range targets {
		sendSSETo(p.Uuid, string(state))
	}
}

func (r *Room) GameStateFor(viewerUuid string) ginGameState {
	playerInfos := r.GameState.PlayerInfo
	if len(playerInfos) == 0 {
		playerInfos = make([]PlayerInfo, len(r.Player))
		for i, p := range r.Player {
			playerInfos[i] = PlayerInfo{
				Uuid:            p.Uuid,
				Name:            p.Name,
				Online:          hasSSEClient(p.Uuid),
				Index:           i,
				CharactersGroup: p.CharactersGroup,
				Character:       p.Character,
				Ready:           p.Ready,
			}
		}
	}

	players := make([]ginPlayerInfo, 0, len(playerInfos))
	myInfo := ginPlayerInfo{}
	for _, p := range playerInfos {
		view := ginPlayerInfo{
			Uuid:            p.Uuid,
			Name:            p.Name,
			Online:          hasSSEClient(p.Uuid),
			Index:           p.Index,
			CharactersGroup: p.CharactersGroup,
			Character:       p.Character,
			Ready:           p.Ready,
			Discarded:       append([]string{}, p.Discarded...),
			Lock:            append([]string{}, p.Lock...),
			Score:           p.Score,
			HandsCount:      len(p.Hands),
		}
		if p.Uuid == viewerUuid {
			view.Hands = append([]string{}, p.Hands...)
			view.New = p.New
			myInfo = view
			continue
		}
		players = append(players, view)
	}

	return ginGameState{
		Playing:         r.Playing,
		RoomRule:        r.RoomRule,
		WallCount:       len(r.GameState.Wall),
		GoldenTile:      r.GameState.GoldenTile,
		PlayerInfo:      players,
		MyInfo:          myInfo,
		CurrentUser:     r.GameState.CurrentUser,
		WaitingResponse: r.GameState.WaitingResponse,
		Out:             r.GameState.Out,
	}
}

func (r *Room) GetPlayerInfo(uuid string) *PlayerInfo {
	for i := range r.GameState.PlayerInfo {
		if r.GameState.PlayerInfo[i].Uuid == uuid {
			return &r.GameState.PlayerInfo[i]
		}
	}
	return nil
}

func (r *Room) Start() {
	if r.Playing || len(r.Player) < 4 {
		return
	}
	r.Starting = false
	r.Playing = true
	rand.Seed(time.Now().UnixNano())
	r.GameState = GameState{}
	r.GameState.Wall = r.buildWall()

	// 发牌
	rule := RuleList[r.RoomRule.Rule]
	r.GameState.PlayerInfo = make([]PlayerInfo, 4)
	for i := 0; i < 4; i++ {
		r.Player[i].Playing = true
		r.GameState.PlayerInfo[i] = PlayerInfo{
			Uuid:            r.Player[i].Uuid,
			Name:            r.Player[i].Name,
			Online:          true,
			Index:           i,
			CharactersGroup: r.Player[i].CharactersGroup,
			Character:       r.Player[i].Character,
			Ready:           r.Player[i].Ready,
			Hands:           make([]string, 0, rule.Total),
		}
		for j := 0; j < rule.Total; j++ {
			r.GameState.PlayerInfo[i].Hands = append(r.GameState.PlayerInfo[i].Hands, r.GameState.Wall[0])
			r.GameState.Wall = r.GameState.Wall[1:]
		}
		sortTiles(r.GameState.PlayerInfo[i].Hands)
	}
	r.GameState.CurrentUser = 0
	r.Draw()
	r.Broadcast()
}

func (r *Room) CanStart() bool {
	if r.Playing || r.Starting || len(r.Player) != 4 {
		return false
	}
	for _, p := range r.Player {
		if !p.Ready {
			return false
		}
	}
	return true
}

func (r *Room) StartCountdown() (bool, int) {
	if !r.CanStart() {
		return false, 15
	}
	r.Starting = true
	go func() {
		for i := 3; i > 0; i-- {
			r.BroadcastEvent(map[string]any{
				"type":    "log",
				"level":   "info",
				"message": fmt.Sprintf("游戏将在 %d 秒后开始", i),
			})
			time.Sleep(1 * time.Second)
		}
		r.Start()
	}()
	return true, 0
}

func (r *Room) buildWall() []string {
	rule := RuleList[r.RoomRule.Rule]
	tiles := []string{
		"Pin1", "Pin2", "Pin3", "Pin4", "Pin5", "Pin6", "Pin7", "Pin8", "Pin9",
		"Sou1", "Sou2", "Sou3", "Sou4", "Sou5", "Sou6", "Sou7", "Sou8", "Sou9",
		"Man1", "Man2", "Man3", "Man4", "Man5", "Man6", "Man7", "Man8", "Man9",
	}
	if rule.Winds {
		tiles = append(tiles, "Ton", "Nan", "Shaa", "Pei")
	}
	if rule.Chun {
		tiles = append(tiles, "Chun")
	}
	if rule.Hatsu {
		tiles = append(tiles, "Hatsu")
	}
	if rule.Haku {
		tiles = append(tiles, "Haku")
	}
	if rule.Season {
		tiles = append(tiles, "spring", "summer", "autumn", "winter")
	}
	if rule.Flowers {
		tiles = append(tiles, "plum", "orchid", "bamboo", "chrysanthemum")
	}

	if rule.Golden && len(tiles) > 0 {
		goldenIndex := rand.Intn(len(tiles))
		r.GameState.GoldenTile = tiles[goldenIndex]
		tiles[goldenIndex] = "Golden"
	}

	wall := append([]string{}, tiles...)
	wall = append(wall, wall...)
	wall = append(wall, wall...)
	rand.Shuffle(len(wall), func(i, j int) {
		wall[i], wall[j] = wall[j], wall[i]
	})
	return wall
}

func (r *Room) Draw() {
	if len(r.GameState.Wall) == 0 {
		r.EndGameReady()
		return
	}
	p := &r.GameState.PlayerInfo[r.GameState.CurrentUser]
	p.New = r.GameState.Wall[0]
	r.GameState.Wall = r.GameState.Wall[1:]
	r.Broadcast()
	if r.RoomRule.SkipOffline && !hasSSEClient(p.Uuid) {
		go func(index int) {
			time.Sleep(100 * time.Millisecond)
			if r.Playing && r.GameState.CurrentUser == index && r.GameState.PlayerInfo[index].New != "" {
				r.Discard(&r.GameState.PlayerInfo[index], len(r.GameState.PlayerInfo[index].Hands))
			}
		}(p.Index)
	}
}

func (r *Room) Discard(pInfo *PlayerInfo, selec int) (bool, int) {
	if r.GameState.CurrentUser != pInfo.Index {
		return false, 61
	}
	if selec < 0 || selec > len(pInfo.Hands) {
		return false, 62
	}

	if selec == len(pInfo.Hands) {
		if pInfo.New == "" {
			return false, 62
		}
		r.GameState.Out = pInfo.New
		pInfo.New = ""
	} else {
		r.GameState.Out = pInfo.Hands[selec]
		pInfo.Hands = append(pInfo.Hands[:selec], pInfo.Hands[selec+1:]...)
		if pInfo.New != "" {
			pInfo.Hands = append(pInfo.Hands, pInfo.New)
			pInfo.New = ""
		}
	}
	sortTiles(pInfo.Hands)
	pInfo.Discarded = append(pInfo.Discarded, r.GameState.Out)

	r.GameState.WaitingResponse = true
	r.GameState.ActionQueue = nil
	r.Broadcast()

	// 启动 2s 定时器
	go func() {
		time.Sleep(2 * time.Second)
		r.ProcessActions()
	}()
	return true, 0
}

func (r *Room) ProcessActions() {
	if !r.GameState.WaitingResponse {
		return
	}
	r.GameState.WaitingResponse = false
	if len(r.GameState.ActionQueue) > 0 {
		sort.Slice(r.GameState.ActionQueue, func(i, j int) bool {
			return r.GameState.ActionQueue[i].Pri < r.GameState.ActionQueue[j].Pri
		})
		best := r.GameState.ActionQueue[0]
		r.ExecuteAction(best)
	} else {
		r.GameState.Out = ""
		r.GameState.CurrentUser = (r.GameState.CurrentUser + 1) % 4
		r.Draw()
	}
}

func (r *Room) ExecuteAction(a Action) {
	p := &r.GameState.PlayerInfo[a.Player]
	actionLabel := actionName(a.Type)
	actionTiles := []string{}
	switch a.Type {
	case 1: // Chow
		tiles := []string{r.GameState.Out}
		for _, i := range a.Selec {
			tiles = append(tiles, p.Hands[i])
		}
		actionTiles = append(actionTiles, tiles...)
		p.Lock = append(p.Lock, tiles...)
		sort.Ints(a.Selec)
		p.Hands = removeIndices(p.Hands, a.Selec...)
		sortTiles(p.Hands)
		r.GameState.CurrentUser = a.Player
		r.GameState.Out = ""
	case 2: // Pong
		tiles := []string{r.GameState.Out, p.Hands[a.Selec[0]], p.Hands[a.Selec[1]]}
		actionTiles = append(actionTiles, tiles...)
		p.Lock = append(p.Lock, tiles...)
		sort.Ints(a.Selec)
		p.Hands = removeIndices(p.Hands, a.Selec...)
		sortTiles(p.Hands)
		r.GameState.CurrentUser = a.Player
		r.GameState.Out = ""
	case 3: // Kong
		tiles := make([]string, 0, len(a.Selec)+1)
		if r.GameState.CurrentUser != a.Player {
			tiles = append(tiles, r.GameState.Out)
		}
		for _, i := range a.Selec {
			tiles = append(tiles, p.Hands[i])
		}
		actionTiles = append(actionTiles, tiles...)
		p.Lock = append(p.Lock, tiles...)
		sort.Ints(a.Selec)
		p.Hands = removeIndices(p.Hands, a.Selec...)
		if p.New != "" {
			p.Hands = append(p.Hands, p.New)
			p.New = ""
		}
		sortTiles(p.Hands)
		r.GameState.CurrentUser = a.Player
		r.GameState.Out = ""
		r.Draw()
		r.BroadcastAction(actionLabel, actionTiles, p.Uuid)
		return
	case 4: // Hu
		actionTiles = append(actionTiles, p.Hands...)
		if r.GameState.CurrentUser == a.Player {
			if p.New != "" {
				actionTiles = append(actionTiles, p.New)
			}
		} else if r.GameState.Out != "" {
			actionTiles = append(actionTiles, r.GameState.Out)
		}
		r.EndGameReady()
		r.BroadcastAction(actionLabel, actionTiles, p.Uuid)
		return
	}
	r.Broadcast()
	r.BroadcastAction(actionLabel, actionTiles, p.Uuid)
}

func actionName(actionType int) string {
	switch actionType {
	case 1:
		return "chow"
	case 2:
		return "pong"
	case 3:
		return "kong"
	case 4:
		return "hu"
	default:
		return ""
	}
}

func (r *Room) BroadcastAction(action string, tiles []string, uuid string) {
	if action == "" {
		return
	}
	r.BroadcastEvent(map[string]any{
		"type":   "action",
		"action": action,
		"tiles":  tiles,
		"uuid":   uuid,
	})
}

func removeIndices(tiles []string, indices ...int) []string {
	sort.Ints(indices)
	res := make([]string, 0, len(tiles)-len(indices))
	idxMap := make(map[int]bool)
	for _, i := range indices {
		idxMap[i] = true
	}
	for i, t := range tiles {
		if !idxMap[i] {
			res = append(res, t)
		}
	}
	return res
}

func validIndices(length int, indices ...int) bool {
	seen := make(map[int]bool, len(indices))
	for _, i := range indices {
		if i < 0 || i >= length || seen[i] {
			return false
		}
		seen[i] = true
	}
	return true
}

func sortTiles(tiles []string) {
	sort.Slice(tiles, func(i, j int) bool {
		return SortRule[tiles[i]] < SortRule[tiles[j]]
	})
}

func (r *Room) canRespondToOut(p *PlayerInfo) (bool, int) {
	if !r.GameState.WaitingResponse || r.GameState.Out == "" {
		return false, 25
	}
	if r.GameState.CurrentUser == p.Index {
		return false, 26
	}
	if isGoldenTile(r.GameState.Out) {
		return false, 27
	}
	return true, 0
}

func (r *Room) huPriority(playerIndex int) int {
	order, ok := HuPri[r.GameState.CurrentUser]
	if !ok {
		return 4
	}
	for pri, index := range order {
		if index == playerIndex {
			return pri + 1
		}
	}
	return 4
}

func (r *Room) hasQueuedAction(playerIndex int) bool {
	for _, action := range r.GameState.ActionQueue {
		if action.Player == playerIndex {
			return true
		}
	}
	return false
}

func (r *Room) EndGameReady() {
	r.Playing = false
	r.Starting = false
	r.GameState = GameState{}
	for _, p := range r.Player {
		p.Playing = false
		p.Ready = true
	}
	r.Broadcast()
}

type ginGameState struct {
	Playing         bool            `json:"Playing"`
	RoomRule        RoomRule        `json:"RoomRule"`
	WallCount       int             `json:"WallCount"`
	GoldenTile      string          `json:"GoldenTile"`
	PlayerInfo      []ginPlayerInfo `json:"PlayerInfo"`
	MyInfo          ginPlayerInfo   `json:"MyInfo"`
	CurrentUser     int             `json:"CurrentUser"`
	WaitingResponse bool            `json:"WaitingResponse"`
	Out             string          `json:"Out"`
}

type ginPlayerInfo struct {
	Uuid            string   `json:"Uuid"`
	Name            string   `json:"Name"`
	Online          bool     `json:"Online"`
	Index           int      `json:"Index"`
	CharactersGroup string   `json:"CharactersGroup"`
	Character       string   `json:"Character"`
	Ready           bool     `json:"Ready"`
	Discarded       []string `json:"Discarded"`
	Lock            []string `json:"Lock"`
	Hands           []string `json:"Hands,omitempty"`
	New             string   `json:"New,omitempty"`
	HandsCount      int      `json:"HandsCount"`
	Score           int      `json:"Score"`
}
