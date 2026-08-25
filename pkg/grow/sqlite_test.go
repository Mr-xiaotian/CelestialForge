package grow

import (
	"path/filepath"
	"testing"
)

// TestLifecycleSQLiteEventRoundTrip 验证事件和父边写入后可被完整读回。
func TestLifecycleSQLiteEventRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lifecycle.sqlite3")
	db, err := OpenLifecycleSQLite(dbPath)
	if err != nil {
		t.Fatalf("OpenLifecycleSQLite() error = %v", err)
	}
	defer db.Close()

	if err := InsertLifecycleEvent(db, LifecycleEventRecord{
		EventID:   1,
		EventType: "seed",
		Plot:      "source",
		TS:        1.0,
	}, nil); err != nil {
		t.Fatalf("InsertLifecycleEvent(seed) error = %v", err)
	}

	if err := InsertLifecycleEvent(db, LifecycleEventRecord{
		EventID:   2,
		EventType: "fruit",
		Plot:      "source",
		TS:        2.0,
	}, []int{1}); err != nil {
		t.Fatalf("InsertLifecycleEvent(fruit) error = %v", err)
	}

	loadedEvent, loadErr := LoadLifecycleEvent(db, 2)
	if loadErr != nil {
		t.Fatalf("LoadLifecycleEvent() error = %v", loadErr)
	}
	if loadedEvent.EventType != "fruit" {
		t.Fatalf("LoadLifecycleEvent() event type = %q, want %q", loadedEvent.EventType, "fruit")
	}

	parentIDs, parentErr := LoadLifecycleEventParents(db, 2)
	if parentErr != nil {
		t.Fatalf("LoadLifecycleEventParents() error = %v", parentErr)
	}
	if len(parentIDs) != 1 || parentIDs[0] != 1 {
		t.Fatalf("LoadLifecycleEventParents() = %v, want [1]", parentIDs)
	}
}

// TestLifecycleSQLiteStatusRoundTrip 验证状态快照的写入、更新和晋升。
func TestLifecycleSQLiteStatusRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lifecycle.sqlite3")
	db, err := OpenLifecycleSQLite(dbPath)
	if err != nil {
		t.Fatalf("OpenLifecycleSQLite() error = %v", err)
	}
	defer db.Close()

	for _, record := range []LifecycleEventRecord{
		{EventID: 1, EventType: "seed", Plot: "stage_a", TS: 1.0},
		{EventID: 2, EventType: "replant", Plot: "stage_a", TS: 2.0},
		{EventID: 3, EventType: "fruit", Plot: "stage_a", TS: 3.0},
	} {
		if err := InsertLifecycleEvent(db, record, nil); err != nil {
			t.Fatalf("InsertLifecycleEvent(%d) error = %v", record.EventID, err)
		}
	}

	if err := UpsertLifecycleStatus(db, LifecycleStatusRecord{
		InputEventID:   1,
		CurrentEventID: 1,
		TaskJSON:       `{"value":"alpha"}`,
		Plot:           "stage_a",
		Status:         "pending",
		ResultJSON:     "null",
		TS:             1.0,
	}); err != nil {
		t.Fatalf("UpsertLifecycleStatus() error = %v", err)
	}

	if err := UpdateLifecycleStatusCurrentEvent(db, 1, 2, "retrying", 2.0); err != nil {
		t.Fatalf("UpdateLifecycleStatusCurrentEvent() error = %v", err)
	}

	if err := PromoteLifecycleStatusSuccess(db, 1, 3, `{"ok":true}`, 3.0); err != nil {
		t.Fatalf("PromoteLifecycleStatusSuccess() error = %v", err)
	}

	loadedStatus, loadErr := LoadLifecycleStatus(db, 1)
	if loadErr != nil {
		t.Fatalf("LoadLifecycleStatus() error = %v", loadErr)
	}
	if loadedStatus.CurrentEventID != 3 {
		t.Fatalf("LoadLifecycleStatus() current event = %d, want %d", loadedStatus.CurrentEventID, 3)
	}
	if loadedStatus.Status != "success" {
		t.Fatalf("LoadLifecycleStatus() status = %q, want %q", loadedStatus.Status, "success")
	}
	if loadedStatus.ResultJSON != `{"ok":true}` {
		t.Fatalf("LoadLifecycleStatus() result json = %q, want %q", loadedStatus.ResultJSON, `{"ok":true}`)
	}
}
