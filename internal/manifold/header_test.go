package manifold

import "testing"

func TestHeaderAllocateRelease(t *testing.T) {
	h, err := NewHeader("hdr-1", "mf-primary", "v-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Allocate(500); err != nil {
		t.Fatal(err)
	}
	if err := h.Allocate(100); err == nil {
		t.Fatal("double allocate should fail")
	}
	h.Release()
	if h.Allocated {
		t.Fatal("expected released")
	}
}

func TestHeaderTablePickAvailable(t *testing.T) {
	h1, _ := NewHeader("h1", "mf-a", "v1", 5)
	h2, _ := NewHeader("h2", "mf-a", "v2", 10)
	tbl := NewHeaderTable(h1, h2)
	picked, err := tbl.PickAvailable("mf-a")
	if err != nil {
		t.Fatal(err)
	}
	if picked.ID != "h2" {
		t.Fatalf("expected higher priority header, got %s", picked.ID)
	}
}
