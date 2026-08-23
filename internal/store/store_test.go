package store

import (
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/model"
)

func TestScheduleCloneIsolation(t *testing.T) {
	ss := NewScheduleStore(NewMemory())
	s := model.BoostSchedule{ID: "sch1", Entries: []model.BoostScheduleEntry{{ID: "e1"}}}
	ss.Save(s)
	clone, err := ss.SnapshotClone("sch1")
	if err != nil {
		t.Fatal(err)
	}
	clone.Entries[0].ID = "changed"
	got, _ := ss.SnapshotClone("sch1")
	if got.Entries[0].ID != "e1" {
		t.Fatal("store keeps isolated clone")
	}
}

func TestActiveEntry(t *testing.T) {
	ss := NewScheduleStore(NewMemory())
	now := time.Unix(100, 0)
	s := model.BoostSchedule{Entries: []model.BoostScheduleEntry{{Start: now.Add(-time.Minute), End: now.Add(time.Minute)}}}
	if _, ok := ss.ActiveEntry(s, now); !ok {
		t.Fatal("active entry")
	}
}