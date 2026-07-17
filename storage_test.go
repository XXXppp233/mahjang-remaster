package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestStoreCreatesSchemaAndTracksOnline(t *testing.T) {
	path := filepath.Join(t.TempDir(), databaseFilename)
	t.Setenv("MAHJONG_DB_PATH", path)
	store, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	p := &Player{Uuid: "player-1", Name: "Alice", IP: "127.0.0.1"}
	if err := store.SetOnline(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM Onlines`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("online count = %d, want 1", count)
	}
	if err := store.DeleteOnline(context.Background(), p.Uuid); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM Onlines`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("online count = %d, want 0", count)
	}
}

func TestOnlineNameIsIndividuallyUnique(t *testing.T) {
	path := filepath.Join(t.TempDir(), databaseFilename)
	t.Setenv("MAHJONG_DB_PATH", path)
	store, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	if err := store.SetOnline(ctx, &Player{Uuid: "a", Name: "Alice", IP: "1"}); err != nil {
		t.Fatal(err)
	}
	err = store.SetOnline(ctx, &Player{Uuid: "b", Name: "alice", IP: "2"})
	if err == nil {
		t.Fatal("expected case-insensitive unique name constraint")
	}
	if err == sql.ErrNoRows {
		t.Fatal("unexpected query error")
	}
}

func TestReplayExportPreservesDatabaseOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), databaseFilename)
	t.Setenv("MAHJONG_DB_PATH", path)
	store, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	room := &Room{Id: "abcdefghijklmnop", Name: "Replay", RoomRule: RoomRule{Rule: 0, SkipOffline: true, MaxWaiting: 21}}
	for i, name := range []string{"A", "B", "C", "D"} {
		room.Player = append(room.Player, &Player{Uuid: "u" + string(rune('1'+i)), Name: name, CharactersGroup: "G", Character: "C"})
	}
	room.GameState.GoldenTile = "Pei"
	ctx := context.Background()
	if err := store.StartGameLog(ctx, room, 666, "FY-XorShift32-v1"); err != nil {
		t.Fatal(err)
	}
	if err := store.LogAction(ctx, room.Id, "u1", "Discard", "Pin1"); err != nil {
		t.Fatal(err)
	}
	replay, err := store.ExportReplay(ctx, room.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Events) != 8 {
		t.Fatalf("event count = %d, want 8", len(replay.Events))
	}
	for i, event := range replay.Events {
		if event.Seq != int64(i+1) {
			t.Fatalf("event %d has seq %d", i, event.Seq)
		}
	}
	if replay.Events[6].Action != "Golden" || replay.Events[7].Action != "Discard" {
		t.Fatalf("unexpected event order: %#v", replay.Events)
	}
}
