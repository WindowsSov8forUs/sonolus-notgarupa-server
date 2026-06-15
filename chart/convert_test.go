package chart

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-core-go/core/resource"
)

func TestGarupaChartConvertToSonolus(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"BPM","beat":0,"value":120},
		{"type":"SV","beat":0.5,"value":1.25,"timingGroup":1},
		{"type":"Single","beat":1,"lane":3,"width":2},
		{"type":"Skill","beat":2,"lane":4,"width":3},
		{"type":"Directional","beat":3,"lane":5,"width":2,"direction":"Left"},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Single","beat":4,"lane":2,"width":2},
			{"type":"Hidden","beat":5,"lane":3,"width":1},
			{"type":"UnknownFutureType","beat":6,"lane":4,"width":3},
			{"type":"Flick","beat":7,"lane":5,"width":2}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Flick","beat":8,"lane":2,"width":2},
			{"type":"Flick","beat":9,"lane":3,"width":3},
			{"type":"Directional","beat":10,"lane":4,"width":3,"direction":"Right"}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Directional","beat":11,"lane":3,"width":2,"direction":"Left"},
			{"type":"Single","beat":12,"lane":4,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Single","beat":13,"lane":2,"width":1},
			{"type":"Directional","beat":14,"lane":3,"width":3,"direction":"Right"},
			{"type":"Single","beat":15,"lane":4,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Hidden","beat":16,"lane":1,"width":1},
			{"type":"Single","beat":17,"lane":2,"width":1},
			{"type":"Single","beat":18,"lane":3,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Single","beat":19,"lane":2,"width":1},
			{"type":"Single","beat":20,"lane":3,"width":1},
			{"type":"Hidden","beat":21,"lane":4,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Hidden","beat":22,"lane":1,"width":1},
			{"type":"Hidden","beat":23,"lane":2,"width":1},
			{"type":"Hidden","beat":24,"lane":3,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Hidden","beat":25,"lane":1,"width":1},
			{"type":"Directional","beat":26,"lane":3,"width":2,"direction":"Right"},
			{"type":"Hidden","beat":27,"lane":5,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Skill","beat":28,"lane":3,"width":1},
			{"type":"Single","beat":29,"lane":4,"width":1}
		]},
		{"type":"Slide","timingGroup":0,"connections":[
			{"type":"Single","beat":30,"lane":3,"width":1},
			{"type":"Skill","beat":31,"lane":4,"width":1}
		]}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	levelData, err := chart.ConvertToSonolus()
	if err != nil {
		t.Fatal(err)
	}

	assertEntity(t, levelData.Entities, resource.EngineArchetypeNameBPMChange, 0, 120)
	assertEntity(t, levelData.Entities, archetypeTapNote, 1, 0)
	assertEntityValues(t, levelData.Entities, archetypeTapNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 2,
	})
	assertEntity(t, levelData.Entities, archetypeSkillNote, 2, 1)
	assertEntityValues(t, levelData.Entities, archetypeSkillNote, 2, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 3,
	})
	assertNoEntity(t, levelData.Entities, archetypeTapNote, 2)
	assertEntity(t, levelData.Entities, archetypeDirectionalFlickNote, 3, 2)
	assertEntity(t, levelData.Entities, archetypeSlideStartNote, 4, -1)
	assertEntityValues(t, levelData.Entities, archetypeSlideStartNote, 4, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 2,
	})
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 5, 0)
	assertEntityValues(t, levelData.Entities, archetypeIgnoredNote, 5, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 1,
	})
	assertEntity(t, levelData.Entities, archetypeSlideTickNote, 6, 1)
	assertEntityValues(t, levelData.Entities, archetypeSlideTickNote, 6, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 3,
	})
	assertEntity(t, levelData.Entities, archetypeSlideEndFlickNote, 7, 2)
	assertEntityValues(t, levelData.Entities, archetypeSlideEndFlickNote, 7, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 2,
	})
	assertEntity(t, levelData.Entities, archetypeSlideStartFlickNote, 8, -1)
	assertEntityValues(t, levelData.Entities, archetypeSlideStartFlickNote, 8, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 2,
	})
	assertEntity(t, levelData.Entities, archetypeSlideTickFlickNote, 9, 0)
	assertEntityValues(t, levelData.Entities, archetypeSlideTickFlickNote, 9, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 3,
	})
	assertEntity(t, levelData.Entities, archetypeSlideEndDirectionalNote, 10, 1)
	assertEntityValues(t, levelData.Entities, archetypeSlideEndDirectionalNote, 10, map[resource.EngineArchetypeDataName]float64{
		dataNameDirection: 1,
		dataNameSize:      3,
		dataNameLong:      0,
	})
	assertEntity(t, levelData.Entities, archetypeSlideStartDirectionalNote, 11, 0)
	assertEntityValues(t, levelData.Entities, archetypeSlideStartDirectionalNote, 11, map[resource.EngineArchetypeDataName]float64{
		dataNameDirection: -1,
		dataNameSize:      2,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartDirectionalNote, 11, map[resource.EngineArchetypeDataName]float64{
		dataNameHeadDirection: -1,
		dataNameHeadSize:      2,
	})
	assertEntity(t, levelData.Entities, archetypeSlideTickDirectionalNote, 14, 0)
	assertEntityValues(t, levelData.Entities, archetypeSlideTickDirectionalNote, 14, map[resource.EngineArchetypeDataName]float64{
		dataNameDirection: 1,
		dataNameSize:      3,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideTickDirectionalNote, 14, map[resource.EngineArchetypeDataName]float64{
		dataNameHeadDirection: 1,
		dataNameHeadSize:      3,
	})
	assertConnectorHeadValuesAbsent(t, levelData.Entities, archetypeSlideStartFlickNote, 8)
	assertConnectorHeadValuesAbsent(t, levelData.Entities, archetypeSlideTickFlickNote, 9)
	assertConnectorModeAbsentBetween(t, levelData.Entities, archetypeSlideStartFlickNote, 8, archetypeSlideTickFlickNote, 9)
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 16, -2)
	assertEntity(t, levelData.Entities, archetypeSlideTickNote, 17, -1)
	assertConnectorBetween(t, levelData.Entities, archetypeIgnoredNote, 16, archetypeSlideTickNote, 17)
	assertConnectorModeBetween(t, levelData.Entities, archetypeIgnoredNote, 16, archetypeSlideTickNote, 17, slideConnectorModeLeadingIgnored)
	assertConnectorModeAbsentBetween(t, levelData.Entities, archetypeSlideTickNote, 17, archetypeSlideEndNote, 18)
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 21, 1)
	assertEntity(t, levelData.Entities, archetypeSlideTickNote, 20, 0)
	assertConnectorBetween(t, levelData.Entities, archetypeSlideTickNote, 20, archetypeIgnoredNote, 21)
	assertConnectorModeAbsentBetween(t, levelData.Entities, archetypeSlideStartNote, 19, archetypeSlideTickNote, 20)
	assertConnectorModeBetween(t, levelData.Entities, archetypeSlideTickNote, 20, archetypeIgnoredNote, 21, slideConnectorModeTrailingIgnored)
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 22, -2)
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 23, -1)
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 24, 0)
	assertConnectorBetween(t, levelData.Entities, archetypeIgnoredNote, 22, archetypeIgnoredNote, 23)
	assertConnectorBetween(t, levelData.Entities, archetypeIgnoredNote, 23, archetypeIgnoredNote, 24)
	assertConnectorModeBetween(t, levelData.Entities, archetypeIgnoredNote, 22, archetypeIgnoredNote, 23, slideConnectorModeAllIgnored)
	assertConnectorModeBetween(t, levelData.Entities, archetypeIgnoredNote, 23, archetypeIgnoredNote, 24, slideConnectorModeAllIgnored)
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 25, -2)
	assertEntity(t, levelData.Entities, archetypeSlideTickDirectionalNote, 26, 0)
	assertEntityValues(t, levelData.Entities, archetypeSlideTickDirectionalNote, 26, map[resource.EngineArchetypeDataName]float64{
		dataNameDirection: 1,
		dataNameSize:      2,
	})
	assertEntity(t, levelData.Entities, archetypeIgnoredNote, 27, 2)
	assertConnectorBetween(t, levelData.Entities, archetypeIgnoredNote, 25, archetypeSlideTickDirectionalNote, 26)
	assertConnectorBetween(t, levelData.Entities, archetypeSlideTickDirectionalNote, 26, archetypeIgnoredNote, 27)
	assertConnectorModeBetween(t, levelData.Entities, archetypeIgnoredNote, 25, archetypeSlideTickDirectionalNote, 26, slideConnectorModeLeadingIgnored)
	assertConnectorModeBetween(t, levelData.Entities, archetypeSlideTickDirectionalNote, 26, archetypeIgnoredNote, 27, slideConnectorModeTrailingIgnored)
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideTickDirectionalNote, 26, map[resource.EngineArchetypeDataName]float64{
		dataNameHeadDirection: 1,
		dataNameHeadSize:      2,
	})
	assertNoEntity(t, levelData.Entities, archetypeDirectionalFlickNote, 26)
	assertEntity(t, levelData.Entities, archetypeSlideStartSkillNote, 28, 0)
	assertNoEntity(t, levelData.Entities, archetypeSlideStartNote, 28)
	assertConnectorBetween(t, levelData.Entities, archetypeSlideStartSkillNote, 28, archetypeSlideEndNote, 29)
	assertEntity(t, levelData.Entities, archetypeSlideEndSkillNote, 31, 1)
	assertNoEntity(t, levelData.Entities, archetypeSlideEndNote, 31)
	assertConnectorBetween(t, levelData.Entities, archetypeSlideStartNote, 30, archetypeSlideEndSkillNote, 31)
	assertEntityValues(t, levelData.Entities, archetypeScrollVelocity, 0.5, map[resource.EngineArchetypeDataName]float64{
		dataNameValue: 1.25,
		dataNameGroup: 1,
	})
	assertEntityValues(t, levelData.Entities, archetypeTapNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 0,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartNote, 4, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 0,
	})
}

func TestHabahiroSlideConnectorSize(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"BPM","beat":0,"value":120},
		{"type":"Slide","connections":[
			{"type":"Single","beat":1,"lane":3,"width":2},
			{"type":"Single","beat":2,"lane":4,"width":1}
		]},
		{"type":"Slide","connections":[
			{"type":"Single","beat":3,"lane":3,"width":2},
			{"type":"Directional","beat":4,"lane":4,"width":3,"direction":"Right"},
			{"type":"Single","beat":5,"lane":5,"width":1}
		]},
		{"type":"Slide","connections":[
			{"type":"Single","beat":6,"lane":3,"width":0},
			{"type":"Single","beat":7,"lane":4,"width":1}
		]},
		{"type":"Slide","connections":[
			{"type":"Directional","beat":8,"lane":3,"width":3,"direction":"Right"},
			{"type":"Single","beat":9,"lane":4,"width":2}
		]},
		{"type":"Slide","connections":[
			{"type":"Directional","beat":10,"lane":3,"width":4,"direction":"Left"},
			{"type":"Directional","beat":11,"lane":4,"width":3,"direction":"Right"}
		]},
		{"type":"Slide","connections":[
			{"type":"Hidden","beat":12,"lane":3,"width":3},
			{"type":"Single","beat":13,"lane":4,"width":1}
		]},
		{"type":"Slide","connections":[
			{"type":"Hidden","beat":14,"lane":3,"width":2},
			{"type":"Hidden","beat":15,"lane":4,"width":3}
		]}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	levelData, err := chart.ConvertToSonolus()
	if err != nil {
		t.Fatal(err)
	}

	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 2,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideTickDirectionalNote, 4, map[resource.EngineArchetypeDataName]float64{
		dataNameSize:          2,
		dataNameHeadDirection: 1,
		dataNameHeadSize:      3,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartNote, 6, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 1,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartDirectionalNote, 8, map[resource.EngineArchetypeDataName]float64{
		dataNameSize:          2,
		dataNameHeadDirection: 1,
		dataNameHeadSize:      3,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartDirectionalNote, 10, map[resource.EngineArchetypeDataName]float64{
		dataNameSize:          1,
		dataNameHeadDirection: -1,
		dataNameHeadSize:      4,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeIgnoredNote, 12, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 3,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeIgnoredNote, 14, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 2,
	})
}

func TestNonHabahiroSlideConnectorSizeIsOne(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"BPM","beat":0,"value":120},
		{"type":"Slide","connections":[
			{"type":"Single","beat":1,"lane":3,"width":1},
			{"type":"Single","beat":2,"lane":4,"width":1}
		]}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	levelData, err := chart.ConvertToSonolus()
	if err != nil {
		t.Fatal(err)
	}

	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameSize: 1,
	})
}

func TestTimingGroupConversion(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"BPM","beat":0,"value":120},
		{"type":"Single","beat":1,"lane":2,"timingGroup":"#A"},
		{"type":"Flick","beat":1,"lane":3,"timingGroup":"#A"},
		{"type":"Directional","beat":1,"lane":4,"direction":"Right","timingGroup":"#B"},
		{"type":"Slide","timingGroup":"#B","connections":[
			{"type":"Single","beat":2,"lane":2},
			{"type":"Single","beat":3,"lane":3}
		]},
		{"type":"SV","beat":0.25,"value":2},
		{"type":"SV","beat":0.5,"value":0,"timingGroup":"#A"},
		{"type":"SV","beat":0.75,"value":-1,"timingGroup":"#B"}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	levelData, err := chart.ConvertToSonolus()
	if err != nil {
		t.Fatal(err)
	}

	assertEntityValues(t, levelData.Entities, archetypeTapNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 1,
	})
	assertEntityValues(t, levelData.Entities, archetypeFlickNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 1,
	})
	assertEntityValues(t, levelData.Entities, archetypeDirectionalFlickNote, 1, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 2,
	})
	assertEntityValues(t, levelData.Entities, archetypeSlideStartNote, 2, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 2,
	})
	assertConnectorHeadValues(t, levelData.Entities, archetypeSlideStartNote, 2, map[resource.EngineArchetypeDataName]float64{
		dataNameGroup: 2,
	})
	assertEntityValues(t, levelData.Entities, archetypeScrollVelocity, 0.25, map[resource.EngineArchetypeDataName]float64{
		dataNameValue: 2,
		dataNameGroup: 0,
	})
	assertEntityValues(t, levelData.Entities, archetypeScrollVelocity, 0.5, map[resource.EngineArchetypeDataName]float64{
		dataNameValue: 0,
		dataNameGroup: 1,
	})
	assertEntityValues(t, levelData.Entities, archetypeScrollVelocity, 0.75, map[resource.EngineArchetypeDataName]float64{
		dataNameValue: -1,
		dataNameGroup: 2,
	})
	assertSimLine(t, levelData.Entities, archetypeTapNote, archetypeFlickNote, 1, 1)
	assertSimLine(t, levelData.Entities, archetypeFlickNote, archetypeDirectionalFlickNote, 1, 0)
	assertSimLineCount(t, levelData.Entities, archetypeFlickNote, archetypeDirectionalFlickNote, 1, 1)
}

func TestScrollVelocityOutputOrder(t *testing.T) {
	var chart GarupaChart
	if err := json.Unmarshal([]byte(`[
		{"type":"BPM","beat":0,"value":120},
		{"type":"SV","beat":2,"value":20,"timingGroup":"#A"},
		{"type":"SV","beat":1,"value":10,"timingGroup":"#B"},
		{"type":"SV","beat":1,"value":11},
		{"type":"SV","beat":0.5,"value":5},
		{"type":"SV","beat":1,"value":12,"timingGroup":"#A"},
		{"type":"SV","beat":1,"value":13},
		{"type":"SV","beat":2,"value":21,"timingGroup":"#B"}
	]`), &chart); err != nil {
		t.Fatal(err)
	}

	levelData, err := chart.ConvertToSonolus()
	if err != nil {
		t.Fatal(err)
	}

	got := scrollVelocitySequence(t, levelData.Entities)
	want := []scrollVelocityRecord{
		{Beat: 0.5, Value: 5, Group: 0},
		{Beat: 1, Value: 10, Group: 2},
		{Beat: 1, Value: 12, Group: 1},
		{Beat: 1, Value: 11, Group: 0},
		{Beat: 1, Value: 13, Group: 0},
		{Beat: 2, Value: 20, Group: 1},
		{Beat: 2, Value: 21, Group: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("scroll velocity count=%d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("scroll velocity[%d]=%#v, want %#v; full=%#v", i, got[i], want[i], got)
		}
	}
}

func TestTimingGroupCapacity(t *testing.T) {
	tooManyGroups := GarupaChart{}
	for i := 1; i <= MaxTimingGroups; i++ {
		tooManyGroups = append(tooManyGroups, &GarupaNote{
			BaseGarupaNote: BaseGarupaNote{
				BaseGarupaObject: BaseGarupaObject{Beat: float64(i), TimingGroup: TimingGroupID("#G" + strconv.Itoa(i))},
				Type:             "Single",
				Lane:             3,
			},
		})
	}
	if _, err := tooManyGroups.ConvertToSonolus(); err == nil {
		t.Fatal("expected too many timing groups error")
	}

	tooManySV := GarupaChart{}
	for i := 0; i <= MaxScrollVelocities; i++ {
		tooManySV = append(tooManySV, &GarupaSV{Beat: float64(i), Value: 1, TimingGroup: GlobalTimingGroupID})
	}
	if _, err := tooManySV.ConvertToSonolus(); err == nil {
		t.Fatal("expected too many scroll velocities error")
	}
}

type scrollVelocityRecord struct {
	Beat  float64
	Value float64
	Group float64
}

func scrollVelocitySequence(t *testing.T, entities []resource.LevelDataEntity) []scrollVelocityRecord {
	t.Helper()
	records := []scrollVelocityRecord{}
	for _, entity := range entities {
		if entity.Archetype != archetypeScrollVelocity {
			continue
		}
		records = append(records, scrollVelocityRecord{
			Beat:  entityValue(t, entity, resource.EngineArchetypeDataNameBeat),
			Value: entityValue(t, entity, dataNameValue),
			Group: entityValue(t, entity, dataNameGroup),
		})
	}
	return records
}

func entityValue(t *testing.T, entity resource.LevelDataEntity, name resource.EngineArchetypeDataName) float64 {
	t.Helper()
	for _, data := range entity.Data {
		valueData, ok := data.(resource.LevelDataEntityValueData)
		if ok && valueData.Name == name {
			return valueData.Value
		}
	}
	t.Fatalf("missing value %q on entity %#v", name, entity)
	return 0
}

func assertEntity(t *testing.T, entities []resource.LevelDataEntity, archetype resource.EngineArchetypeName, beat float64, laneOrBPM float64) {
	t.Helper()
	for _, entity := range entities {
		if entity.Archetype != archetype {
			continue
		}
		if !hasValue(entity, resource.EngineArchetypeDataNameBeat, beat) {
			continue
		}
		if archetype == resource.EngineArchetypeNameBPMChange {
			if hasValue(entity, resource.EngineArchetypeDataNameBPM, laneOrBPM) {
				return
			}
			continue
		}
		if hasValue(entity, dataNameLane, laneOrBPM) {
			return
		}
	}
	t.Fatalf("missing entity archetype=%q beat=%v value=%v in %#v", archetype, beat, laneOrBPM, entities)
}

func assertEntityValues(t *testing.T, entities []resource.LevelDataEntity, archetype resource.EngineArchetypeName, beat float64, values map[resource.EngineArchetypeDataName]float64) {
	t.Helper()
	for _, entity := range entities {
		if entity.Archetype != archetype {
			continue
		}
		if !hasValue(entity, resource.EngineArchetypeDataNameBeat, beat) {
			continue
		}
		for name, value := range values {
			if !hasValue(entity, name, value) {
				t.Fatalf("missing data %q=%v on entity %#v", name, value, entity)
			}
		}
		return
	}
	t.Fatalf("missing entity archetype=%q beat=%v in %#v", archetype, beat, entities)
}

func assertConnectorHeadValues(t *testing.T, entities []resource.LevelDataEntity, archetype resource.EngineArchetypeName, beat float64, values map[resource.EngineArchetypeDataName]float64) {
	t.Helper()
	headRef := entityRef(t, entities, archetype, beat)
	connector := connectorWithHeadRef(t, entities, headRef)
	for name, value := range values {
		if !hasValue(connector, name, value) {
			t.Fatalf("missing connector data %q=%v on %#v", name, value, connector)
		}
	}
}

func assertConnectorHeadValuesAbsent(t *testing.T, entities []resource.LevelDataEntity, archetype resource.EngineArchetypeName, beat float64) {
	t.Helper()
	headRef := entityRef(t, entities, archetype, beat)
	connector := connectorWithHeadRef(t, entities, headRef)
	if hasDataName(connector, dataNameHeadDirection) || hasDataName(connector, dataNameHeadSize) {
		t.Fatalf("unexpected directional head data on connector %#v", connector)
	}
}

func entityRef(t *testing.T, entities []resource.LevelDataEntity, archetype resource.EngineArchetypeName, beat float64) string {
	t.Helper()
	for _, entity := range entities {
		if entity.Archetype != archetype || !hasValue(entity, resource.EngineArchetypeDataNameBeat, beat) {
			continue
		}
		if entity.Name == "" {
			t.Fatalf("entity archetype=%q beat=%v has no ref name: %#v", archetype, beat, entity)
		}
		return entity.Name
	}
	t.Fatalf("missing entity ref archetype=%q beat=%v in %#v", archetype, beat, entities)
	return ""
}

func connectorWithHeadRef(t *testing.T, entities []resource.LevelDataEntity, headRef string) resource.LevelDataEntity {
	t.Helper()
	for _, entity := range entities {
		if entity.Archetype != archetypeStraightSlideConnector && entity.Archetype != archetypeCurvedSlideConnector {
			continue
		}
		if hasRef(entity, dataNameHead, headRef) {
			return entity
		}
	}
	t.Fatalf("missing connector with head ref %q in %#v", headRef, entities)
	return resource.LevelDataEntity{}
}

func assertConnectorBetween(t *testing.T, entities []resource.LevelDataEntity, headArchetype resource.EngineArchetypeName, headBeat float64, tailArchetype resource.EngineArchetypeName, tailBeat float64) {
	t.Helper()
	connectorBetween(t, entities, headArchetype, headBeat, tailArchetype, tailBeat)
}

func assertConnectorModeBetween(t *testing.T, entities []resource.LevelDataEntity, headArchetype resource.EngineArchetypeName, headBeat float64, tailArchetype resource.EngineArchetypeName, tailBeat float64, mode int) {
	t.Helper()
	connector := connectorBetween(t, entities, headArchetype, headBeat, tailArchetype, tailBeat)
	if !hasValue(connector, dataNameMode, float64(mode)) {
		t.Fatalf("missing connector mode=%v on %#v", mode, connector)
	}
}

func assertConnectorModeAbsentBetween(t *testing.T, entities []resource.LevelDataEntity, headArchetype resource.EngineArchetypeName, headBeat float64, tailArchetype resource.EngineArchetypeName, tailBeat float64) {
	t.Helper()
	connector := connectorBetween(t, entities, headArchetype, headBeat, tailArchetype, tailBeat)
	if hasDataName(connector, dataNameMode) {
		t.Fatalf("unexpected connector mode on %#v", connector)
	}
}

func connectorBetween(t *testing.T, entities []resource.LevelDataEntity, headArchetype resource.EngineArchetypeName, headBeat float64, tailArchetype resource.EngineArchetypeName, tailBeat float64) resource.LevelDataEntity {
	t.Helper()
	headRef := entityRef(t, entities, headArchetype, headBeat)
	tailRef := entityRef(t, entities, tailArchetype, tailBeat)
	for _, entity := range entities {
		if entity.Archetype != archetypeStraightSlideConnector && entity.Archetype != archetypeCurvedSlideConnector {
			continue
		}
		if hasRef(entity, dataNameHead, headRef) && hasRef(entity, dataNameTail, tailRef) {
			return entity
		}
	}
	t.Fatalf("missing connector head=%s/%v tail=%s/%v in %#v", headArchetype, headBeat, tailArchetype, tailBeat, entities)
	return resource.LevelDataEntity{}
}

func assertSimLine(t *testing.T, entities []resource.LevelDataEntity, aArchetype resource.EngineArchetypeName, bArchetype resource.EngineArchetypeName, beat float64, group float64) {
	t.Helper()
	aRef := entityRef(t, entities, aArchetype, beat)
	bRef := entityRef(t, entities, bArchetype, beat)
	for _, entity := range entities {
		if entity.Archetype != archetypeSimLine {
			continue
		}
		if hasRef(entity, dataNameA, aRef) && hasRef(entity, dataNameB, bRef) && hasValue(entity, dataNameGroup, group) {
			return
		}
	}
	t.Fatalf("missing simline a=%s b=%s beat=%v group=%v in %#v", aArchetype, bArchetype, beat, group, entities)
}

func assertSimLineCount(t *testing.T, entities []resource.LevelDataEntity, aArchetype resource.EngineArchetypeName, bArchetype resource.EngineArchetypeName, beat float64, want int) {
	t.Helper()
	aRef := entityRef(t, entities, aArchetype, beat)
	bRef := entityRef(t, entities, bArchetype, beat)
	count := 0
	for _, entity := range entities {
		if entity.Archetype == archetypeSimLine && hasRef(entity, dataNameA, aRef) && hasRef(entity, dataNameB, bRef) {
			count++
		}
	}
	if count != want {
		t.Fatalf("simline count a=%s b=%s beat=%v is %d, want %d in %#v", aArchetype, bArchetype, beat, count, want, entities)
	}
}

func assertNoEntity(t *testing.T, entities []resource.LevelDataEntity, archetype resource.EngineArchetypeName, beat float64) {
	t.Helper()
	for _, entity := range entities {
		if entity.Archetype == archetype && hasValue(entity, resource.EngineArchetypeDataNameBeat, beat) {
			t.Fatalf("unexpected entity archetype=%q beat=%v: %#v", archetype, beat, entity)
		}
	}
}

func assertNoEntityAtBeat(t *testing.T, entities []resource.LevelDataEntity, beat float64) {
	t.Helper()
	for _, entity := range entities {
		if hasValue(entity, resource.EngineArchetypeDataNameBeat, beat) {
			t.Fatalf("unexpected entity at beat=%v: %#v", beat, entity)
		}
	}
}

func hasValue(entity resource.LevelDataEntity, name resource.EngineArchetypeDataName, value float64) bool {
	for _, data := range entity.Data {
		valueData, ok := data.(resource.LevelDataEntityValueData)
		if !ok {
			continue
		}
		if valueData.Name == name && valueData.Value == value {
			return true
		}
	}
	return false
}

func hasRef(entity resource.LevelDataEntity, name resource.EngineArchetypeDataName, ref string) bool {
	for _, data := range entity.Data {
		refData, ok := data.(resource.LevelDataEntityRefData)
		if !ok {
			continue
		}
		if refData.Name == name && refData.Ref == ref {
			return true
		}
	}
	return false
}

func hasDataName(entity resource.LevelDataEntity, name resource.EngineArchetypeDataName) bool {
	for _, data := range entity.Data {
		switch typedData := data.(type) {
		case resource.LevelDataEntityValueData:
			if typedData.Name == name {
				return true
			}
		case resource.LevelDataEntityRefData:
			if typedData.Name == name {
				return true
			}
		}
	}
	return false
}
