package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const roomIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

func newRoomID() string {
	b := make([]byte, 16)
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = roomIDAlphabet[int(raw[i])%len(roomIDAlphabet)]
	}
	return string(b)
}

//go:embed FrontEnd/dist/*
var frontend embed.FS

var (
	Rooms      = make(map[string]*Room)
	Players    = make(map[string]*Player)
	RoomsMu    sync.RWMutex
	PlayersMu  sync.RWMutex
	ServerKeys = make(map[string]*rsa.PrivateKey)
	SSEClients = make(map[string]map[chan string]struct{}) // [uuid]: channels
	SSEMu      sync.RWMutex
	Database   *Store
)

func main() {
	var err error
	Database, err = OpenStore()
	if err != nil {
		panic("unable to open state database: " + err.Error())
	}
	defer Database.Close()

	r := gin.Default()

	// API Routes
	api := r.Group("/api")
	{
		api.GET("/handshake", handleHandshake)
		api.GET("/config", handleConfig)
		api.POST("/login", handleLogin)
		api.POST("/register", handleRegister)

		auth := api.Group("/")
		auth.Use(AuthMiddleware())
		{
			auth.POST("/logout", handleLogout)
			auth.GET("/stream", handleStream)
			auth.GET("/rooms", handleRooms)
			auth.GET("/replays/:roomid/download", handleReplayDownload)
			auth.POST("/rooms/create", handleCreateRoom)
			auth.POST("/rooms/join", handleJoinRoom)
			auth.POST("/rooms/leave", handleLeaveRoom)
			auth.POST("/rooms/chat", handleRoomChat)
			auth.POST("/rooms/user", handleRoomUser)
			auth.POST("/rooms/start", handleStartGame)
			auth.POST("/gaming/:action", handleGameAction)
			auth.POST("", handleGameAction)
		}
	}

	// Serve Frontend
	feFS, _ := fs.Sub(frontend, "FrontEnd/dist")
	r.NoRoute(func(c *gin.Context) {
		// setFrontendCrossOriginPolicy(c)

		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		fPath := strings.TrimPrefix(path, "/")
		if fPath == "" {
			fPath = "index.html"
		}

		if _, err := fs.Stat(feFS, fPath); err == nil {
			http.FileServer(http.FS(feFS)).ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback
		indexData, err := fs.ReadFile(feFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend build not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	port := getEnv("PORT", "8080")
	r.Run(":" + port)
}

// func characterAssetOrigins() []string {
// 	seen := map[string]bool{}
// 	origins := []string{}
// 	for _, rawURL := range GlobalConf.CharactersMap {
// 		u, err := url.Parse(rawURL)
// 		if err != nil || u.Scheme == "" || u.Host == "" {
// 			continue
// 		}
// 		if u.Scheme != "http" && u.Scheme != "https" {
// 			continue
// 		}
// 		origin := u.Scheme + "://" + u.Host
// 		if !seen[origin] {
// 			seen[origin] = true
// 			origins = append(origins, origin)
// 		}
// 	}
// 	return origins
// }

// func setFrontendCrossOriginPolicy(c *gin.Context) {
// 	connectSrc := "'self'"
// 	imgSrc := "'self' data: blob:"
// 	mediaSrc := "'self' blob:"
// 	for _, origin := range characterAssetOrigins() {
// 		imgSrc += " " + origin
// 		mediaSrc += " " + origin
// 	}
// 	c.Writer.Header().Set(
// 		"Content-Security-Policy",
// 		"default-src 'self'; "+
// 			"script-src 'self'; "+
// 			"style-src 'self' 'unsafe-inline'; "+
// 			"font-src 'self' data:; "+
// 			"connect-src "+connectSrc+"; "+
// 			"img-src "+imgSrc+"; "+
// 			"media-src "+mediaSrc,
// 	)
// }

func handleHandshake(c *gin.Context) {
	ip := c.ClientIP()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate key"})
		return
	}
	ServerKeys[ip] = key

	pubKey := &key.PublicKey
	pubASN1, _ := x509.MarshalPKIXPublicKey(pubKey)
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	c.String(http.StatusOK, string(pubBytes))
}

func playerFromCookie(cookie string, ip string) (*Player, bool) {
	PlayersMu.RLock()
	player, ok := Players[cookie]
	PlayersMu.RUnlock()
	if ok {
		player.IP = ip
		return player, true
	}
	return nil, false
}

func playerFromRequest(c *gin.Context) (*Player, bool) {
	if GlobalConf.LoginRequire {
		if cookie, err := c.Cookie("session"); err == nil {
			if player, ok := playerFromCookie(cookie, c.ClientIP()); ok {
				return player, true
			}
		}
		return nil, false
	}

	playerID := c.GetHeader("X-Player-UUID")
	if playerID == "" {
		playerID = c.Query("player")
	}
	if playerID == "" {
		return nil, false
	}
	PlayersMu.RLock()
	player, ok := Players[playerID]
	PlayersMu.RUnlock()
	if ok {
		player.IP = c.ClientIP()
	}
	return player, ok
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		player, ok := playerFromRequest(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("player", player)
		c.Next()
	}
}

func registerSSEClient(playerUuid string, ch chan string) {
	SSEMu.Lock()
	defer SSEMu.Unlock()

	if SSEClients[playerUuid] == nil {
		SSEClients[playerUuid] = make(map[chan string]struct{})
	}
	SSEClients[playerUuid][ch] = struct{}{}
}

func unregisterSSEClient(playerUuid string, ch chan string) bool {
	SSEMu.Lock()
	defer SSEMu.Unlock()

	clients := SSEClients[playerUuid]
	if clients == nil {
		return false
	}
	delete(clients, ch)
	if len(clients) == 0 {
		delete(SSEClients, playerUuid)
		return false
	}
	return true
}

func hasSSEClient(playerUuid string) bool {
	SSEMu.RLock()
	defer SSEMu.RUnlock()

	return len(SSEClients[playerUuid]) > 0
}

func sseClientCount(playerUuid string) int {
	SSEMu.RLock()
	defer SSEMu.RUnlock()
	return len(SSEClients[playerUuid])
}

func sendSSETo(playerUuid string, msg string) {
	SSEMu.RLock()
	defer SSEMu.RUnlock()

	for ch := range SSEClients[playerUuid] {
		select {
		case ch <- msg:
		default:
			// Buffer full or slow client.
		}
	}
}

func handleStream(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)
	_ = Database.SetOnline(c.Request.Context(), p)

	ch := make(chan string, 100)
	registerSSEClient(p.Uuid, ch)

	defer func() {
		hasConnections := unregisterSSEClient(p.Uuid, ch)
		if !hasConnections {
			_ = Database.DeleteOnline(context.Background(), p.Uuid)
		}

		if !GlobalConf.LoginRequire && !hasConnections {
			PlayersMu.Lock()
			if p.Room != nil {
				room := p.Room
				if success, _ := room.RemoveUser(p); success {
					_ = Database.DeletePlayer(context.Background(), p.Uuid)
					if len(room.Player) == 0 {
						RoomsMu.Lock()
						delete(Rooms, room.Id)
						RoomsMu.Unlock()
						_ = Database.DeleteRoom(context.Background(), room.Id)
					} else {
						_ = Database.SaveRoom(context.Background(), room)
						room.BroadcastEvent(gin.H{
							"type": "room_user",
							"uuid": p.Uuid,
							"left": true,
						})
					}
				} else { // 所有玩家离线则解散房间
					roomPlayer := len(room.Player)
					for _, p := range room.Player {
						if !p.Online {
							roomPlayer--
						}
					}
					if roomPlayer == 0 {
						RoomsMu.Lock()
						delete(Rooms, room.Id)
						RoomsMu.Unlock()
					}
				}
			}
			delete(Players, p.Uuid)
			PlayersMu.Unlock()
		}
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	if p.Room != nil {
		go p.Room.Broadcast()
	}

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true
		case <-ticker.C:
			c.SSEvent("heartbeat", "ping")
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func userResponse(p *Player) gin.H {
	if p == nil {
		return nil
	}
	roomID := ""
	if p.Room != nil {
		roomID = p.Room.Id
	}
	return gin.H{
		"Uuid": p.Uuid,
		"Name": p.Name,
		"Room": roomID,
		"Grade": gin.H{
			"1st": p.Grade.First,
			"2nd": p.Grade.Second,
			"3rd": p.Grade.Third,
			"4th": p.Grade.Fourth,
		},
		"IP": p.IP,
	}
}

func handleConfig(c *gin.Context) {
	var player *Player
	if p, ok := playerFromRequest(c); ok {
		player = p
	}

	res := gin.H{
		"ShowIPLocation": GlobalConf.ShowIPLocation,
		"LoginRequire":   GlobalConf.LoginRequire,
		"User":           userResponse(player),
		"CharactersMap":  GlobalConf.CharactersMap,
	}

	c.JSON(http.StatusOK, res)
}

func handleLogin(c *gin.Context) {
	var req struct {
		Account  string `json:"Account"`
		Password string `json:"Password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if req.Account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Account is required"})
		return
	}
	if strings.EqualFold(strings.TrimSpace(req.Account), "Server") {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Server is a reserved name"})
		return
	}

	var playerUuid string
	var player *Player

	if !GlobalConf.LoginRequire {
		playerUuid = uuid.New().String()
		player = &Player{
			Uuid: playerUuid,
			Name: req.Account,
			IP:   c.ClientIP(),
			Grade: Grade{
				First:  0,
				Second: 0,
				Third:  0,
				Fourth: 0,
			},
		}
		PlayersMu.Lock()
		Players[playerUuid] = player
		PlayersMu.Unlock()
		if err := Database.SetOnline(c.Request.Context(), player); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to record online player"})
			return
		}
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"status": false, "message": "Login with database is currently not supported"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"User":   userResponse(player),
	})
}

func handleRegister(c *gin.Context) {
	handleLogin(c)
}

func handleCreateRoom(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	if p.Room != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": "您有未完成的对局"})
		return
	}

	var req struct {
		Name string `json:"Name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	room := &Room{
		Name: req.Name,
		Id:   newRoomID(),
		RoomRule: RoomRule{
			Rule:       0,
			MaxWaiting: 10,
		},
	}
	if !GlobalConf.LoginRequire {
		room.RoomRule.SkipOffline = true
	}
	room.AddUser(p)
	p.Room = room

	RoomsMu.Lock()
	Rooms[room.Id] = room
	RoomsMu.Unlock()
	if err := Database.SaveRoom(c.Request.Context(), room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to persist room"})
		return
	}
	_ = Database.SaveLobbyPlayer(c.Request.Context(), room.Id, p)

	c.JSON(http.StatusOK, roomJoinResponse(room, p))
}

func handleJoinRoom(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	roomID := c.Query("roomid")
	// if roomID == "" {
	// 	var req struct {
	// 		Id   string `json:"Id"`
	// 		Name string `json:"Name"`
	// 	}
	// 	if err := c.ShouldBindJSON(&req); err == nil {
	// 		roomID = req.Id
	// 		if roomID == "" {
	// 			roomID = req.Name
	// 		}
	// 	}
	// }
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if p.Room != nil && p.Room.Id != roomID {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": "您有未完成的对局"})
		return
	}

	RoomsMu.RLock()
	room, ok := Rooms[roomID]
	RoomsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Room not found"})
		return
	}

	success, errCode := room.AddUser(p)
	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Failed to join", "code": errCode})
		return
	}
	p.Room = room
	_ = Database.SaveRoom(c.Request.Context(), room)
	_ = Database.SaveLobbyPlayer(c.Request.Context(), room.Id, p)
	if info := room.GetPlayerInfo(p.Uuid); info != nil {
		info.Online = true
	}
	room.BroadcastEvent(gin.H{
		"type":  "room_user",
		"uuid":  p.Uuid,
		"name":  p.Name,
		"ready": p.Ready,
		"Owner": roomOwner(room),
		"decorator": gin.H{
			"org":   p.CharactersGroup,
			"chara": p.Character,
		},
	})

	c.JSON(http.StatusOK, roomJoinResponse(room, p))
}

func handleLeaveRoom(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	leavePlayer(c, p, c.Query("roomid"))
}

func leavePlayer(c *gin.Context, p *Player, roomID string) {
	if p.Room == nil || (roomID != "" && p.Room.Id != roomID) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Not in this room"})
		return
	}

	room := p.Room
	if info := room.GetPlayerInfo(p.Uuid); info != nil {
		info.Online = false
	}
	if room.Playing || p.Playing {
		c.JSON(http.StatusOK, gin.H{"status": true})
		room.Broadcast()
		return
	}

	room.RemoveUser(p)
	_ = Database.DeletePlayer(c.Request.Context(), p.Uuid)
	_ = Database.SaveRoom(c.Request.Context(), room)
	if len(room.Player) == 0 {
		RoomsMu.Lock()
		delete(Rooms, room.Id)
		RoomsMu.Unlock()
		_ = Database.DeleteRoom(c.Request.Context(), room.Id)
	} else {
		room.BroadcastEvent(gin.H{
			"type":  "room_user",
			"uuid":  p.Uuid,
			"left":  true,
			"Owner": roomOwner(room),
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func handleRoomChat(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	roomID := c.Query("roomid")
	if roomID == "" && p.Room != nil {
		roomID = p.Room.Id
	}
	if queryUuid := c.Query("uuid"); queryUuid != "" && queryUuid != p.Uuid {
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": "Invalid uuid"})
		return
	}
	if p.Room == nil || p.Room.Id != roomID {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Not in this room"})
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
		return
	}

	p.Room.BroadcastEvent(gin.H{
		"type":    "chat",
		"uuid":    p.Uuid,
		"name":    p.Name,
		"message": req.Message,
	})
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func handleRoomUser(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	switch c.Query("action") {
	case "ready":
		handleRoomUserReady(c, p)
	case "unready":
		handleRoomUserUnready(c, p)
	case "leave":
		leavePlayer(c, p, "")
	case "kick":
		handleRoomUserKick(c, p)
	case "update_rule":
		handleRoomUserUpdateRule(c, p)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid action"})
	}
}

func handleRoomUserUnready(c *gin.Context, p *Player) {
	if p.Room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Not in a room"})
		return
	}
	if p.Room.Playing || p.Room.Starting {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Game is starting or already started"})
		return
	}

	p.Ready = false
	if info := p.Room.GetPlayerInfo(p.Uuid); info != nil {
		info.Ready = false
	}
	_ = Database.SaveLobbyPlayer(c.Request.Context(), p.Room.Id, p)
	_ = Database.SaveRoom(c.Request.Context(), p.Room)

	p.Room.BroadcastEvent(gin.H{
		"type":  "room_user",
		"uuid":  p.Uuid,
		"name":  p.Name,
		"ready": p.Ready,
		"Owner": roomOwner(p.Room),
		"decorator": gin.H{
			"org":   p.CharactersGroup,
			"chara": p.Character,
		},
	})
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func handleRoomUserKick(c *gin.Context, p *Player) {
	if p.Room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Not in a room"})
		return
	}
	if len(p.Room.Player) == 0 || p.Room.Player[0].Uuid != p.Uuid {
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": "Only owner can kick"})
		return
	}

	targetUuid := c.Query("target")
	var targetPlayer *Player
	for _, tp := range p.Room.Player {
		if tp.Uuid == targetUuid {
			targetPlayer = tp
			break
		}
	}

	if targetPlayer == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Target player not found in room"})
		return
	}

	room := p.Room
	room.RemoveUser(targetPlayer)
	_ = Database.DeletePlayer(c.Request.Context(), targetPlayer.Uuid)
	_ = Database.SaveRoom(c.Request.Context(), room)
	reason := "您已被房主踢出房间"
	room.BroadcastEvent(gin.H{
		"type":   "room_user",
		"uuid":   targetUuid,
		"left":   true,
		"reason": reason,
		"Owner":  roomOwner(room),
	})
	c.JSON(http.StatusOK, gin.H{"status": true, "message": reason})
}

func handleRoomUserUpdateRule(c *gin.Context, p *Player) {
	if p.Room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Not in a room"})
		return
	}
	if len(p.Room.Player) == 0 || p.Room.Player[0].Uuid != p.Uuid {
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": "Only owner can update rule"})
		return
	}

	var req RoomRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
		return
	}

	success, errCode := p.Room.Update(req)
	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Failed to update rule", "code": errCode})
		return
	}

	p.Room.Broadcast()
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func handleRoomUserReady(c *gin.Context, p *Player) {
	if p.Room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Not in a room"})
		return
	}

	var req struct {
		Decorator struct {
			Org   string `json:"org"`
			Chara string `json:"chara"`
		} `json:"decorator"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
		return
	}
	if req.Decorator.Org == "" || req.Decorator.Chara == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Character is required"})
		return
	}

	p.Ready = true
	p.CharactersGroup = req.Decorator.Org
	p.Character = req.Decorator.Chara
	_ = Database.SaveLobbyPlayer(c.Request.Context(), p.Room.Id, p)
	if info := p.Room.GetPlayerInfo(p.Uuid); info != nil {
		info.Ready = true
		info.CharactersGroup = p.CharactersGroup
		info.Character = p.Character
		_ = Database.SavePlayer(c.Request.Context(), p.Room.Id, *info)
	}
	_ = Database.SaveRoom(c.Request.Context(), p.Room)

	p.Room.BroadcastEvent(gin.H{
		"type":  "room_user",
		"uuid":  p.Uuid,
		"name":  p.Name,
		"ready": p.Ready,
		"Owner": roomOwner(p.Room),
		"decorator": gin.H{
			"org":   p.CharactersGroup,
			"chara": p.Character,
		},
	})
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func handleLogout(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)
	if sseClientCount(p.Uuid) > 1 {
		c.JSON(http.StatusOK, gin.H{"status": true, "sharedSession": true})
		return
	}
	_ = Database.DeleteOnline(c.Request.Context(), p.Uuid)

	if p.Room != nil {
		if info := p.Room.GetPlayerInfo(p.Uuid); info != nil {
			info.Online = false
		}
		if !p.Room.Playing && !p.Playing {
			room := p.Room
			room.RemoveUser(p)
			if len(room.Player) == 0 {
				RoomsMu.Lock()
				delete(Rooms, room.Id)
				RoomsMu.Unlock()
			}
		} else {
			p.Room.Broadcast()
		}
	}

	if p.Cookie != "" {
		PlayersMu.Lock()
		delete(Players, p.Cookie)
		PlayersMu.Unlock()
		c.SetCookie("session", "", -1, "/", "", false, true)
	} else {
		PlayersMu.Lock()
		delete(Players, p.Uuid)
		PlayersMu.Unlock()
	}

	c.JSON(http.StatusOK, gin.H{"status": true})
}

func handleReplayDownload(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)
	roomID := c.Param("roomid")
	if p.Room == nil || p.Room.Id != roomID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Replay does not belong to the current room"})
		return
	}
	replay, err := Database.ExportReplay(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Replay not found"})
		return
	}
	filename := fmt.Sprintf("MahJongLTS-%s-%d.json", roomID, replay.ExportedAt)
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.JSON(http.StatusOK, replay)
}

func handleRooms(c *gin.Context) {
	RoomsMu.RLock()
	defer RoomsMu.RUnlock()

	roomList := make([]gin.H, 0, len(Rooms))
	for _, room := range Rooms {
		roomList = append(roomList, roomResponse(room))
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   true,
		"roomlist": roomList,
	})
}

func roomResponse(room *Room) gin.H {
	return gin.H{
		"Name":      room.Name,
		"Id":        room.Id,
		"Owner":     roomOwner(room),
		"Player":    len(room.Player),
		"Watcher":   len(room.Watcher),
		"Playering": room.Playing,
	}
}

func roomJoinResponse(room *Room, p *Player) gin.H {
	return gin.H{
		"status":   true,
		"Room":     roomResponse(room),
		"State":    room.GameStateFor(p.Uuid),
		"RuleList": ruleListResponse(),
	}
}

func ruleListResponse() []gin.H {
	keys := make([]int, 0, len(RuleList))
	for index := range RuleList {
		keys = append(keys, index)
	}
	sort.Ints(keys)

	rules := make([]gin.H, 0, len(keys))
	for _, index := range keys {
		rules = append(rules, gin.H{
			"Index": index,
			"Name":  RuleNames[index],
		})
	}
	return rules
}

func roomOwner(room *Room) string {
	if room == nil || len(room.Player) == 0 {
		return ""
	}
	return room.Player[0].Uuid
}

func handleStartGame(c *gin.Context) {
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	if p.Room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not in a room"})
		return
	}

	if len(p.Room.Player) == 0 || p.Room.Player[0].Uuid != p.Uuid {
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": "Only owner can start game"})
		return
	}

	success, errCode := p.Room.StartCountdown()
	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Failed to start game", "code": errCode})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Game countdown started"})
}

func handleGameAction(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		action = c.Param("action")
	}
	playerRaw, _ := c.Get("player")
	p := playerRaw.(*Player)

	RoomsMu.RLock()
	room := p.Room
	RoomsMu.RUnlock()

	if room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not in a room"})
		return
	}

	info := room.GetPlayerInfo(p.Uuid)
	if info == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player info not found"})
		return
	}

	var res bool
	var errCode int
	room.mu.Lock()
	defer room.mu.Unlock()

	switch action {
	case "discard":
		var req struct {
			PlayerIndex int `json:"PlayerIndex"`
			Selec       int `json:"Selec"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
			return
		}
		if req.PlayerIndex != info.Index {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": 63})
			return
		}
		res, errCode = room.Discard(info, req.Selec)
	case "chow":
		var req struct {
			PlayerIndex int    `json:"PlayerIndex"`
			Selec       [2]int `json:"Selec"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
			return
		}
		if req.PlayerIndex != info.Index {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": 63})
			return
		}
		res, errCode = room.CheckChow(info, req.Selec)
	case "pong":
		var req struct {
			PlayerIndex int    `json:"PlayerIndex"`
			Selec       [2]int `json:"Selec"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
			return
		}
		if req.PlayerIndex != info.Index {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": 63})
			return
		}
		res, errCode = room.CheckPong(info, req.Selec)
	case "kong":
		var req struct {
			PlayerIndex int   `json:"PlayerIndex"`
			Selec       []int `json:"Selec"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
			return
		}
		if req.PlayerIndex != info.Index {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": 63})
			return
		}
		res, errCode = room.CheckKong(info, req.Selec)
	case "hu":
		var req struct {
			PlayerIndex int `json:"PlayerIndex"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid request"})
			return
		}
		if req.PlayerIndex != info.Index {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": 63})
			return
		}
		res, errCode = room.CheckHu(info)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": 20})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": res, "code": errCode})
}
