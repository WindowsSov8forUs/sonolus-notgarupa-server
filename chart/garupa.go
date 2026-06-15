package chart

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
)

type GarupaChart []GarupaChartObject

const GlobalTimingGroupID TimingGroupID = "#Global"

var timingGroupPattern = regexp.MustCompile(`^#[A-Za-z0-9 -]+$`)

type TimingGroupID string

func (id *TimingGroupID) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*id = normalizeTimingGroupValue(value)
	return nil
}

func (id TimingGroupID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(normalizeTimingGroup(id)))
}

func normalizeTimingGroup(id TimingGroupID) TimingGroupID {
	if id == "" {
		return GlobalTimingGroupID
	}
	if !timingGroupPattern.MatchString(string(id)) {
		return GlobalTimingGroupID
	}
	return id
}

func normalizeTimingGroupValue(value any) TimingGroupID {
	switch value := value.(type) {
	case nil:
		return GlobalTimingGroupID
	case string:
		return normalizeTimingGroup(TimingGroupID(value))
	case float64:
		if value <= 0 || math.Trunc(value) != value {
			return GlobalTimingGroupID
		}
		return TimingGroupID("#" + strconv.FormatInt(int64(value), 10))
	default:
		return GlobalTimingGroupID
	}
}

func (chart *GarupaChart) UnmarshalJSON(data []byte) error {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return err
	}

	objects := make([]GarupaChartObject, 0, len(rawItems))
	for _, rawItem := range rawItems {
		var header struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawItem, &header); err != nil {
			return err
		}

		switch header.Type {
		case "BPM":
			var item GarupaBPM
			if err := json.Unmarshal(rawItem, &item); err != nil {
				return err
			}
			objects = append(objects, &item)
		case "SV":
			var item GarupaSV
			if err := json.Unmarshal(rawItem, &item); err != nil {
				return err
			}
			item.normalize()
			objects = append(objects, &item)
		case "Single", "Flick", "Skill":
			var item GarupaNote
			if err := json.Unmarshal(rawItem, &item); err != nil {
				return err
			}
			item.normalize()
			objects = append(objects, &item)
		case "Directional":
			var item GarupaDirectionalNote
			if err := json.Unmarshal(rawItem, &item); err != nil {
				return err
			}
			item.normalize()
			objects = append(objects, &item)
		case "Slide":
			var item GarupaSlide
			if err := json.Unmarshal(rawItem, &item); err != nil {
				return err
			}
			item.normalize()
			objects = append(objects, &item)
		default:
			// Unknown top-level objects are intentionally ignored for forward compatibility.
		}
	}
	*chart = objects
	return nil
}

func (chart GarupaChart) MarshalJSON() ([]byte, error) {
	fields := make([]map[string]any, 0, len(chart))
	for _, object := range chart {
		data, err := json.Marshal(object)
		if err != nil {
			return nil, err
		}
		var field map[string]any
		if err := json.Unmarshal(data, &field); err != nil {
			return nil, err
		}
		field["type"] = object.garupaType()
		switch object := object.(type) {
		case *GarupaNote:
			normalizeTimingGroupField(field, object.TimingGroup, false)
		case *GarupaDirectionalNote:
			normalizeTimingGroupField(field, object.TimingGroup, false)
		case *GarupaSlide:
			normalizeTimingGroupField(field, object.TimingGroup, false)
			removeSlideConnectionTimingGroups(field)
		case *GarupaSV:
			normalizeTimingGroupField(field, object.TimingGroup, true)
		}
		fields = append(fields, field)
	}
	return json.Marshal(fields)
}

func removeSlideConnectionTimingGroups(field map[string]any) {
	connections, ok := field["connections"].([]any)
	if !ok {
		return
	}
	for _, connection := range connections {
		connectionField, ok := connection.(map[string]any)
		if ok {
			delete(connectionField, "timingGroup")
		}
	}
}

func normalizeTimingGroupField(field map[string]any, id TimingGroupID, keepGlobal bool) {
	normalized := normalizeTimingGroup(id)
	if normalized == GlobalTimingGroupID && !keepGlobal {
		delete(field, "timingGroup")
		return
	}
	field["timingGroup"] = string(normalized)
}

type GarupaChartObject interface {
	garupaType() string
	Convert(ctx context.Context) error
}

type BaseGarupaObject struct {
	Beat        float64       `json:"beat"`
	TimingGroup TimingGroupID `json:"timingGroup,omitempty"`
}

func (object *BaseGarupaObject) normalize() {
	object.TimingGroup = normalizeTimingGroup(object.TimingGroup)
}

type BaseGarupaNote struct {
	BaseGarupaObject
	Type  string  `json:"type"`
	Lane  float64 `json:"lane"`
	Width float64 `json:"width,omitempty"`
}

type GarupaNote struct {
	BaseGarupaNote
}

func (note *GarupaNote) garupaType() string {
	return note.Type
}

type GarupaDirectionalNote struct {
	BaseGarupaNote
	Direction string `json:"direction"`
}

func (note *GarupaDirectionalNote) garupaType() string {
	return "Directional"
}

type GarupaSlideConnection struct {
	BaseGarupaNote
	Direction string `json:"direction,omitempty"`
}

func (connection GarupaSlideConnection) isHidden() bool {
	return connection.Type == "Hidden"
}

func (connection GarupaSlideConnection) isFlickTail() bool {
	return connection.Type == "Flick" || connection.Type == "Directional"
}

type GarupaSlide struct {
	Connections []GarupaSlideConnection `json:"connections"`
	TimingGroup TimingGroupID           `json:"timingGroup,omitempty"`
}

func (slide *GarupaSlide) garupaType() string {
	return "Slide"
}

func (slide *GarupaSlide) normalize() {
	slide.TimingGroup = normalizeTimingGroup(slide.TimingGroup)
	for i := range slide.Connections {
		slide.Connections[i].normalize()
	}
}

type GarupaBPM struct {
	Beat  float64 `json:"beat"`
	Value float64 `json:"value"`
}

func (bpm *GarupaBPM) garupaType() string {
	return "BPM"
}

type GarupaSV struct {
	Beat        float64       `json:"beat"`
	Value       float64       `json:"value"`
	TimingGroup TimingGroupID `json:"timingGroup,omitempty"`
}

func (sv *GarupaSV) garupaType() string {
	return "SV"
}

func (sv *GarupaSV) normalize() {
	sv.TimingGroup = normalizeTimingGroup(sv.TimingGroup)
}
