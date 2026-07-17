package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const databaseFilename = "MahJongLTS.db"

var gameLogTablePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type Store struct {
	db       *sql.DB
	path     string
	mu       sync.Mutex
	logTable map[string]string
}

type ReplayEvent struct {
	Seq       int64  `json:"seq"`
	Timestamp int64  `json:"timestamp"`
	User      string `json:"user"`
	Action    string `json:"action"`
	Value     string `json:"value"`
}

type ReplayExport struct {
	Version    int           `json:"version"`
	RoomID     string        `json:"room_id"`
	ExportedAt int64         `json:"exported_at"`
	Events     []ReplayEvent `json:"events"`
}

func OpenStore() (*Store, error) {
	path, err := selectDatabasePath()
	if err != nil {
		return nil, err
	}
	databaseURL := &url.URL{Scheme: "file", Path: path}
	query := databaseURL.Query()
	query.Add("_pragma", "busy_timeout(5000)")
	query.Add("_pragma", "foreign_keys(1)")
	query.Add("_pragma", "journal_mode(WAL)")
	databaseURL.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", databaseURL.String())
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path, logTable: make(map[string]string)}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	// An SSE connection cannot survive a server restart.
	if _, err := db.Exec(`DELETE FROM Onlines`); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func selectDatabasePath() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("MAHJONG_DB_PATH")); configured != "" {
		dir := filepath.Dir(configured)
		if err := probeWritable(dir); err != nil {
			return "", fmt.Errorf("MAHJONG_DB_PATH directory is not writable: %w", err)
		}
		return configured, nil
	}
	cwd, err := os.Getwd()
	if err == nil && probeWritable(cwd) == nil {
		return filepath.Join(cwd, databaseFilename), nil
	}
	if err := probeWritable(os.TempDir()); err != nil {
		return "", fmt.Errorf("neither current directory nor temporary directory is writable: %w", err)
	}
	return filepath.Join(os.TempDir(), databaseFilename), nil
}

func probeWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".mahjong-write-test-*")
	if err != nil {
		return err
	}
	name := f.Name()
	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(name)
		return closeErr
	}
	return os.Remove(name)
}

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS Onlines (
    UUID TEXT PRIMARY KEY,
    Name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    IP TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS Rooms (
    RoomID TEXT PRIMARY KEY,
    Name TEXT NOT NULL UNIQUE,
    Status TEXT NOT NULL,
    Player1 TEXT,
    Player2 TEXT,
    Player3 TEXT,
    Player4 TEXT,
    CurrentPlayer TEXT,
    CreatedAt INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS Players (
    PlayerID TEXT PRIMARY KEY,
    RoomID TEXT NOT NULL,
    Ready INTEGER NOT NULL DEFAULT 0,
    CGroup TEXT NOT NULL DEFAULT '',
    Chara TEXT NOT NULL DEFAULT '',
    LockTiles TEXT NOT NULL DEFAULT '[]',
    Hands TEXT NOT NULL DEFAULT '[]',
    Discarded TEXT NOT NULL DEFAULT '[]',
    NewTile TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (RoomID) REFERENCES Rooms(RoomID) ON DELETE CASCADE
);`)
	return err
}

func (s *Store) SetOnline(ctx context.Context, p *Player) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO Onlines(UUID, Name, IP) VALUES(?, ?, ?)
ON CONFLICT(UUID) DO UPDATE SET Name=excluded.Name, IP=excluded.IP`, p.Uuid, p.Name, p.IP)
	return err
}

func (s *Store) DeleteOnline(ctx context.Context, uuid string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM Onlines WHERE UUID=?`, uuid)
	return err
}

func (s *Store) SaveRoom(ctx context.Context, room *Room) error {
	players := make([]any, 4)
	for i := 0; i < len(room.Player) && i < 4; i++ {
		players[i] = room.Player[i].Uuid
	}
	status := "waiting"
	if room.Starting {
		status = "starting"
	}
	if room.Playing {
		status = "playing"
	}
	current := ""
	if room.Playing && room.GameState.CurrentUser >= 0 && room.GameState.CurrentUser < len(room.GameState.PlayerInfo) {
		current = room.GameState.PlayerInfo[room.GameState.CurrentUser].Uuid
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO Rooms(RoomID,Name,Status,Player1,Player2,Player3,Player4,CurrentPlayer,CreatedAt)
VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(RoomID) DO UPDATE SET Name=excluded.Name,Status=excluded.Status,
Player1=excluded.Player1,Player2=excluded.Player2,Player3=excluded.Player3,Player4=excluded.Player4,CurrentPlayer=excluded.CurrentPlayer`,
		room.Id, room.Name, status, players[0], players[1], players[2], players[3], current, time.Now().UnixMilli())
	return err
}

func jsonTiles(v []string) string { b, _ := json.Marshal(v); return string(b) }

func (s *Store) SavePlayer(ctx context.Context, roomID string, p PlayerInfo) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO Players(PlayerID,RoomID,Ready,CGroup,Chara,LockTiles,Hands,Discarded,NewTile)
VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(PlayerID) DO UPDATE SET RoomID=excluded.RoomID,Ready=excluded.Ready,
CGroup=excluded.CGroup,Chara=excluded.Chara,LockTiles=excluded.LockTiles,Hands=excluded.Hands,
Discarded=excluded.Discarded,NewTile=excluded.NewTile`, p.Uuid, roomID, p.Ready, p.CharactersGroup,
		p.Character, jsonTiles(p.Lock), jsonTiles(p.Hands), jsonTiles(p.Discarded), p.New)
	return err
}

func (s *Store) SaveLobbyPlayer(ctx context.Context, roomID string, p *Player) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO Players(PlayerID,RoomID,Ready,CGroup,Chara)
VALUES(?,?,?,?,?) ON CONFLICT(PlayerID) DO UPDATE SET RoomID=excluded.RoomID,Ready=excluded.Ready,
CGroup=excluded.CGroup,Chara=excluded.Chara`, p.Uuid, roomID, p.Ready, p.CharactersGroup, p.Character)
	return err
}

func (s *Store) DeleteRoom(ctx context.Context, roomID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM Rooms WHERE RoomID=?`, roomID)
	return err
}

func (s *Store) DeletePlayer(ctx context.Context, playerID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM Players WHERE PlayerID=?`, playerID)
	return err
}

func (s *Store) StartGameLog(ctx context.Context, room *Room, seed uint64, algo string) error {
	table := fmt.Sprintf("GameLog_%s_%d", room.Id, time.Now().UnixMilli())
	if !gameLogTablePattern.MatchString(table) {
		return errors.New("unsafe game log table name")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`CREATE TABLE "` + table + `" (Seq INTEGER PRIMARY KEY AUTOINCREMENT, Timestamp INTEGER NOT NULL, User TEXT NOT NULL, Action TEXT NOT NULL, Value TEXT NOT NULL)`); err != nil {
		return err
	}
	for i, p := range room.Player {
		value := fmt.Sprintf("%s:%s/%s", p.Name, p.CharactersGroup, p.Character)
		if _, err = tx.Exec(`INSERT INTO "`+table+`"(Timestamp,User,Action,Value) VALUES(?,?,?,?)`, i+1, p.Uuid, "Ready", value); err != nil {
			return err
		}
	}
	initial := []struct {
		n             int
		action, value string
	}{
		{5, "SetRule", fmt.Sprintf("Rule:%d;Skip:%t;MaxW:%d", room.RoomRule.Rule, room.RoomRule.SkipOffline, room.RoomRule.MaxWaiting)},
		{6, "Shuffle", fmt.Sprintf("Seed:%d;Algo:%s;Wall:v1", seed, algo)},
	}
	for _, e := range initial {
		if _, err = tx.Exec(`INSERT INTO "`+table+`"(Timestamp,User,Action,Value) VALUES(?,?,?,?)`, e.n, "Server", e.action, e.value); err != nil {
			return err
		}
	}
	if room.GameState.GoldenTile != "" {
		if _, err = tx.Exec(`INSERT INTO "`+table+`"(Timestamp,User,Action,Value) VALUES(7,'Server','Golden',?)`, room.GameState.GoldenTile); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	s.mu.Lock()
	s.logTable[room.Id] = table
	s.mu.Unlock()
	return nil
}

func (s *Store) LogAction(ctx context.Context, roomID, user, action, value string) error {
	s.mu.Lock()
	table := s.logTable[roomID]
	s.mu.Unlock()
	if table == "" || !gameLogTablePattern.MatchString(table) {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO "`+table+`"(Timestamp,User,Action,Value) VALUES(?,?,?,?)`, time.Now().UnixMilli(), user, action, value)
	return err
}

func (s *Store) ExportReplay(ctx context.Context, roomID string) (ReplayExport, error) {
	s.mu.Lock()
	table := s.logTable[roomID]
	s.mu.Unlock()
	if table == "" || !gameLogTablePattern.MatchString(table) {
		return ReplayExport{}, errors.New("replay not found")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT Seq,Timestamp,User,Action,Value FROM "`+table+`" ORDER BY Seq`)
	if err != nil {
		return ReplayExport{}, err
	}
	defer rows.Close()
	result := ReplayExport{Version: 1, RoomID: roomID, ExportedAt: time.Now().UnixMilli(), Events: []ReplayEvent{}}
	for rows.Next() {
		var event ReplayEvent
		if err := rows.Scan(&event.Seq, &event.Timestamp, &event.User, &event.Action, &event.Value); err != nil {
			return ReplayExport{}, err
		}
		result.Events = append(result.Events, event)
	}
	return result, rows.Err()
}

func (s *Store) Close() error { return s.db.Close() }
