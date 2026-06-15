package chart

import (
	"encoding/json"
	"testing"
)

func TestGarupaChartTimingGroupJSON(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"Single","beat":1,"lane":3,"timingGroup":0},
		{"type":"Flick","beat":2,"lane":4,"timingGroup":1},
		{"type":"Directional","beat":3,"lane":5,"direction":"Left","timingGroup":"#A-B"},
		{"type":"Single","beat":4,"lane":2,"timingGroup":"invalid"},
		{"type":"Slide","timingGroup":2,"connections":[
			{"type":"Single","beat":5,"lane":2,"timingGroup":"#IgnoredConnectionGroup"},
			{"type":"Single","beat":6,"lane":3}
		]},
		{"type":"SV","beat":0.5,"value":1.25},
		{"type":"SV","beat":1.5,"value":-2,"timingGroup":"#A-B"}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	if got := chart[0].(*GarupaNote).TimingGroup; got != GlobalTimingGroupID {
		t.Fatalf("numeric global note timingGroup=%q", got)
	}
	if got := chart[1].(*GarupaNote).TimingGroup; got != "#1" {
		t.Fatalf("numeric group note timingGroup=%q", got)
	}
	if got := chart[2].(*GarupaDirectionalNote).TimingGroup; got != "#A-B" {
		t.Fatalf("string group directional timingGroup=%q", got)
	}
	if got := chart[3].(*GarupaNote).TimingGroup; got != GlobalTimingGroupID {
		t.Fatalf("invalid group note timingGroup=%q", got)
	}
	if got := chart[4].(*GarupaSlide).TimingGroup; got != "#2" {
		t.Fatalf("slide timingGroup=%q", got)
	}
	if got := chart[5].(*GarupaSV).TimingGroup; got != GlobalTimingGroupID {
		t.Fatalf("default SV timingGroup=%q", got)
	}

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	var fields []map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}

	assertNoTimingGroupField(t, fields[0])
	assertTimingGroupField(t, fields[1], "#1")
	assertTimingGroupField(t, fields[2], "#A-B")
	assertNoTimingGroupField(t, fields[3])
	assertTimingGroupField(t, fields[4], "#2")
	assertSlideConnectionsHaveNoTimingGroup(t, fields[4])
	assertTimingGroupField(t, fields[5], string(GlobalTimingGroupID))
	assertTimingGroupField(t, fields[6], "#A-B")
}

func TestGarupaChartTimingGroupIndexMap(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"Single","beat":1,"lane":3,"timingGroup":"#A"},
		{"type":"Slide","timingGroup":"#B","connections":[
			{"type":"Single","beat":2,"lane":3},
			{"type":"Single","beat":3,"lane":4}
		]},
		{"type":"SV","beat":1,"value":2,"timingGroup":"#A"},
		{"type":"Directional","beat":4,"lane":3,"direction":"Right"},
		{"type":"SV","beat":2,"value":0,"timingGroup":"#C"}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	groups := chart.TimingGroupIndexMap()
	assertGroupIndex(t, groups, GlobalTimingGroupID, 0)
	assertGroupIndex(t, groups, "#A", 1)
	assertGroupIndex(t, groups, "#B", 2)
	assertGroupIndex(t, groups, "#C", 3)
	if len(groups) != 4 {
		t.Fatalf("group count=%d, want 4: %#v", len(groups), groups)
	}
}

func assertTimingGroupField(t *testing.T, field map[string]any, want string) {
	t.Helper()
	if got := field["timingGroup"]; got != want {
		t.Fatalf("timingGroup=%#v, want %q in %#v", got, want, field)
	}
}

func assertNoTimingGroupField(t *testing.T, field map[string]any) {
	t.Helper()
	if _, ok := field["timingGroup"]; ok {
		t.Fatalf("unexpected timingGroup in %#v", field)
	}
}

func assertSlideConnectionsHaveNoTimingGroup(t *testing.T, field map[string]any) {
	t.Helper()
	connections, ok := field["connections"].([]any)
	if !ok {
		t.Fatalf("missing connections in %#v", field)
	}
	for _, connection := range connections {
		connectionField, ok := connection.(map[string]any)
		if !ok {
			t.Fatalf("invalid connection %#v", connection)
		}
		assertNoTimingGroupField(t, connectionField)
	}
}

func assertGroupIndex(t *testing.T, groups TimingGroupIndexMap, id TimingGroupID, want int) {
	t.Helper()
	if got := groups[id]; got != want {
		t.Fatalf("group %q index=%d, want %d in %#v", id, got, want, groups)
	}
}
