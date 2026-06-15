package chart

import (
	"context"
	"sort"
	"strconv"

	"github.com/WindowsSov8forUs/sonolus-core-go/core/resource"
)

const (
	contextValuesKey = "values"
)

const (
	slideConnectorModeNormal = iota
	slideConnectorModeLeadingIgnored
	slideConnectorModeTrailingIgnored
	slideConnectorModeAllIgnored
)

func (note *GarupaNote) Convert(ctx context.Context) error {
	archetype := archetypeForGarupaNoteType(note.Type)
	group := timingGroupIndex(ctx, note.TimingGroup)
	return appendIntermediate(ctx, &Intermediate{
		Archetype: archetype,
		Data: map[resource.EngineArchetypeDataName]any{
			resource.EngineArchetypeDataNameBeat: note.Beat,
			dataNameLane:                         garupaLaneToEngineLane(note.Lane),
			dataNameSize:                         normalizedWidth(note.Width),
			dataNameGroup:                        group,
		},
		Sim: true,
	})
}

func (note *GarupaDirectionalNote) Convert(ctx context.Context) error {
	group := timingGroupIndex(ctx, note.TimingGroup)
	return appendIntermediate(ctx, &Intermediate{
		Archetype: archetypeDirectionalFlickNote,
		Data: map[resource.EngineArchetypeDataName]any{
			resource.EngineArchetypeDataNameBeat: note.Beat,
			dataNameLane:                         garupaLaneToEngineLane(note.Lane),
			dataNameDirection:                    directionValue(note.Direction),
			dataNameSize:                         normalizedWidth(note.Width),
			dataNameGroup:                        group,
		},
		Sim: true,
	})
}

func (slide *GarupaSlide) Convert(ctx context.Context) error {
	connections := sortedGarupaSlideConnections(slide.Connections)
	return convertGarupaSlide(ctx, connections, timingGroupIndex(ctx, slide.TimingGroup))
}

func (bpm *GarupaBPM) Convert(ctx context.Context) error {
	return appendIntermediate(ctx, &Intermediate{
		Archetype: resource.EngineArchetypeNameBPMChange,
		Data: map[resource.EngineArchetypeDataName]any{
			resource.EngineArchetypeDataNameBeat: bpm.Beat,
			resource.EngineArchetypeDataNameBPM:  bpm.Value,
		},
		Sim: false,
	})
}

func (sv *GarupaSV) Convert(ctx context.Context) error {
	group := timingGroupIndex(ctx, sv.TimingGroup)
	return appendScrollVelocity(ctx, sv.Beat, sv.Value, group)
}

func appendScrollVelocity(ctx context.Context, beat float64, value float64, group int) error {
	return appendIntermediate(ctx, &Intermediate{
		Archetype: archetypeScrollVelocity,
		Data: map[resource.EngineArchetypeDataName]any{
			resource.EngineArchetypeDataNameBeat: beat,
			dataNameValue:                        value,
			dataNameGroup:                        group,
		},
		Sim: false,
	})
}

func archetypeForGarupaNoteType(noteType string) resource.EngineArchetypeName {
	switch noteType {
	case "Flick":
		return archetypeFlickNote
	case "Skill":
		return archetypeSkillNote
	case "Single":
		return archetypeTapNote
	default:
		// Unknown note types currently fall back to Single semantics.
		return archetypeTapNote
	}
}

func garupaLaneToEngineLane(lane float64) float64 {
	return lane - 3.0
}

func normalizedWidth(width float64) float64 {
	if width <= 0 {
		return 1
	}
	return width
}

func directionValue(direction string) float64 {
	if direction == "Left" {
		return -1
	}
	return 1
}

func slideStartArchetype(connection GarupaSlideConnection) resource.EngineArchetypeName {
	switch connection.Type {
	case "Directional":
		return archetypeSlideStartDirectionalNote
	case "Flick":
		return archetypeSlideStartFlickNote
	case "Skill":
		return archetypeSlideStartSkillNote
	default:
		return archetypeSlideStartNote
	}
}

func slideTickArchetype(connection GarupaSlideConnection) resource.EngineArchetypeName {
	switch connection.Type {
	case "Directional":
		return archetypeSlideTickDirectionalNote
	case "Flick":
		return archetypeSlideTickFlickNote
	default:
		return archetypeSlideTickNote
	}
}

func slideEndArchetype(connection GarupaSlideConnection) resource.EngineArchetypeName {
	switch connection.Type {
	case "Directional":
		return archetypeSlideEndDirectionalNote
	case "Flick":
		return archetypeSlideEndFlickNote
	case "Skill":
		return archetypeSlideEndSkillNote
	default:
		return archetypeSlideEndNote
	}
}

func applyDirectionalSlideData(intermediate *Intermediate, connection GarupaSlideConnection) {
	if connection.Type != "Directional" {
		return
	}
	intermediate.Data[dataNameDirection] = directionValue(connection.Direction)
}

func newSlideConnector(archetype resource.EngineArchetypeName, group int, size float64, first, start, head, tail *Intermediate) *Intermediate {
	connector := &Intermediate{
		Archetype: archetype,
		Data: map[resource.EngineArchetypeDataName]any{
			dataNameFirst: first,
			dataNameStart: start,
			dataNameHead:  head,
			dataNameTail:  tail,
			dataNameSize:  size,
			dataNameGroup: group,
		},
		Sim: false,
	}
	appendConnectorHeadData(connector, head)
	return connector
}

func slideConnectorSize(connections []GarupaSlideConnection) float64 {
	for _, connection := range connections {
		if connection.Type == "Directional" {
			continue
		}
		return normalizedWidth(connection.Width)
	}
	return 1
}

func appendConnectorHeadData(connector, head *Intermediate) {
	switch head.Archetype {
	case archetypeSlideStartDirectionalNote, archetypeSlideTickDirectionalNote:
		connector.Data[dataNameHeadDirection] = head.Data[dataNameDirection]
		connector.Data[dataNameHeadSize] = head.Data[dataNameSize]
	}
}

func convertGarupaSlide(ctx context.Context, connections []GarupaSlideConnection, group int) error {
	if len(connections) == 0 {
		return nil
	}

	visibleIndices := slideVisibleIndices(connections)
	connectorArchetype := archetypeStraightSlideConnector
	for _, connection := range connections {
		if connection.isHidden() {
			connectorArchetype = archetypeCurvedSlideConnector
			break
		}
	}

	nodes := make([]*Intermediate, len(connections))
	first := slideFirstAnchor(visibleIndices)
	connectorSize := slideConnectorSize(connections)
	prevVisible := -1
	for i, connection := range connections {
		node := newSlideNode(connection, slideNodeRoleForConnection(i, len(connections), connection), len(visibleIndices), group)
		if first >= 0 && !connection.isHidden() {
			node.Data[dataNameFirst] = nodes[first]
			if prevVisible >= 0 {
				node.Data[dataNamePrev] = nodes[prevVisible]
			} else {
				node.Data[dataNamePrev] = node
			}
		}
		nodes[i] = node
		if !connection.isHidden() {
			prevVisible = i
		}
	}

	if len(visibleIndices) > 0 {
		last := visibleIndices[len(visibleIndices)-1]
		nodes[first].Data[dataNameLast] = nodes[last]
	}

	appends := make([]*Intermediate, 0, len(nodes)+max(0, len(nodes)-1))
	appends = append(appends, nodes...)

	for i := 0; i < len(nodes)-1; i++ {
		firstRef, startRef, endRef := slideConnectorRefs(i, nodes, visibleIndices)
		connector := newSlideConnector(connectorArchetype, group, connectorSize, firstRef, startRef, nodes[i], nodes[i+1])
		connector.Data[dataNameEnd] = endRef
		mode := slideConnectorMode(i, visibleIndices)
		if mode != slideConnectorModeNormal {
			connector.Data[dataNameMode] = float64(mode)
		}
		appends = append(appends, connector)
	}

	for _, intermediate := range appends {
		if err := appendIntermediate(ctx, intermediate); err != nil {
			return err
		}
	}
	return nil
}

func sortedGarupaSlideConnections(connections []GarupaSlideConnection) []GarupaSlideConnection {
	connections = append([]GarupaSlideConnection(nil), connections...)
	sort.Slice(connections, func(i, j int) bool {
		return connections[i].Beat < connections[j].Beat
	})
	return connections
}

func slideVisibleIndices(connections []GarupaSlideConnection) []int {
	indices := make([]int, 0, len(connections))
	for i, connection := range connections {
		if !connection.isHidden() {
			indices = append(indices, i)
		}
	}
	return indices
}

func slideFirstAnchor(visibleIndices []int) int {
	if len(visibleIndices) == 0 {
		return 0
	}
	return visibleIndices[0]
}

func slideConnectorMode(connectorIndex int, visibleIndices []int) int {
	if len(visibleIndices) == 0 {
		return slideConnectorModeAllIgnored
	}
	firstVisibleIndex := visibleIndices[0]
	lastVisibleIndex := visibleIndices[len(visibleIndices)-1]
	if connectorIndex < firstVisibleIndex {
		return slideConnectorModeLeadingIgnored
	}
	if connectorIndex >= lastVisibleIndex {
		return slideConnectorModeTrailingIgnored
	}
	return slideConnectorModeNormal
}

type slideNodeRole int

const (
	slideNodeRoleIgnored slideNodeRole = iota
	slideNodeRoleStart
	slideNodeRoleTick
	slideNodeRoleEnd
)

func slideNodeRoleForConnection(index int, total int, connection GarupaSlideConnection) slideNodeRole {
	if connection.isHidden() {
		return slideNodeRoleIgnored
	}
	if index == 0 {
		return slideNodeRoleStart
	}
	if index == total-1 {
		return slideNodeRoleEnd
	}
	return slideNodeRoleTick
}

func newSlideNode(connection GarupaSlideConnection, role slideNodeRole, visibleCount int, group int) *Intermediate {
	node := &Intermediate{
		Archetype: archetypeIgnoredNote,
		Data: map[resource.EngineArchetypeDataName]any{
			resource.EngineArchetypeDataNameBeat: connection.Beat,
			dataNameLane:                         garupaLaneToEngineLane(connection.Lane),
			dataNameSize:                         normalizedWidth(connection.Width),
			dataNameGroup:                        group,
		},
		Sim: false,
	}
	switch role {
	case slideNodeRoleStart:
		node.Archetype = slideStartArchetype(connection)
		node.Sim = true
	case slideNodeRoleTick:
		node.Archetype = slideTickArchetype(connection)
	case slideNodeRoleEnd:
		node.Archetype = slideEndArchetype(connection)
		node.Sim = true
		if connection.Type == "Directional" {
			node.Data[dataNameLong] = 0.0
		}
		if connection.Type == "Flick" {
			node.Data[dataNameLong] = 0.0
		}
	}
	if role != slideNodeRoleIgnored {
		applyDirectionalSlideData(node, connection)
	}
	if role == slideNodeRoleEnd && (connection.Type == "Directional" || connection.Type == "Flick") && visibleCount == 2 {
		node.Data[dataNameLong] = 0.0
	}
	return node
}

func slideConnectorRefs(connectorIndex int, nodes []*Intermediate, visibleIndices []int) (*Intermediate, *Intermediate, *Intermediate) {
	if len(visibleIndices) == 0 {
		anchor := nodes[0]
		return anchor, anchor, anchor
	}

	firstRef := nodes[visibleIndices[0]]
	startIndex := visibleIndices[0]
	endIndex := visibleIndices[len(visibleIndices)-1]
	for _, visibleIndex := range visibleIndices {
		if visibleIndex <= connectorIndex {
			startIndex = visibleIndex
		}
		if visibleIndex > connectorIndex {
			endIndex = visibleIndex
			break
		}
	}
	return firstRef, nodes[startIndex], nodes[endIndex]
}

type convertContextValue struct {
	Entities             []*resource.LevelDataEntity
	BeatToIntermediates  map[float64][]*Intermediate
	IntermediateToRef    map[*Intermediate]string
	IntermediateToEntity map[*Intermediate]*resource.LevelDataEntity
	TimingGroups         TimingGroupIndexMap
	refCounter           int64
}

func timingGroupIndex(ctx context.Context, id TimingGroupID) int {
	ctxValues := ctx.Value(contextValuesKey).(*convertContextValue)
	return ctxValues.TimingGroups.Index(id)
}

func getRef(ctx context.Context, intermediate *Intermediate) string {
	ctxValues := ctx.Value(contextValuesKey).(*convertContextValue)
	ref, ok := ctxValues.IntermediateToRef[intermediate]
	if !ok {
		ref = strconv.FormatInt(ctxValues.refCounter, 36)
		ctxValues.refCounter++
		ctxValues.IntermediateToRef[intermediate] = ref
		entity, ok := ctxValues.IntermediateToEntity[intermediate]
		if ok {
			entity.Name = ref
			ctxValues.IntermediateToEntity[intermediate] = entity
		}
	}
	return ref
}

func appendIntermediate(ctx context.Context, intermediate *Intermediate) error {
	ctxValues := ctx.Value(contextValuesKey).(*convertContextValue)
	entity := resource.LevelDataEntity{
		Archetype: intermediate.Archetype,
		Data:      []resource.LevelDataEntityData{},
	}

	if intermediate.Sim {
		beat, ok := intermediate.Data[resource.EngineArchetypeDataNameBeat].(float64)
		if !ok {
			return ErrUnexpectedBeat
		}
		ctxValues.BeatToIntermediates[beat] = append(ctxValues.BeatToIntermediates[beat], intermediate)
	}

	if ref, ok := ctxValues.IntermediateToRef[intermediate]; ok {
		entity.Name = ref
	}

	ctxValues.IntermediateToEntity[intermediate] = &entity
	ctxValues.Entities = append(ctxValues.Entities, &entity)

	keys := make([]resource.EngineArchetypeDataName, 0, len(intermediate.Data))
	for key := range intermediate.Data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, key := range keys {
		if valueNumber, ok := intermediate.Data[key].(float64); ok {
			entity.Data = append(entity.Data, resource.LevelDataEntityValueData{
				Name:  key,
				Value: valueNumber,
			})
			continue
		}
		if valueNumber, ok := intermediate.Data[key].(int); ok {
			entity.Data = append(entity.Data, resource.LevelDataEntityValueData{
				Name:  key,
				Value: float64(valueNumber),
			})
			continue
		}
		if valueIntermediate, ok := intermediate.Data[key].(*Intermediate); ok {
			entity.Data = append(entity.Data, resource.LevelDataEntityRefData{
				Name: key,
				Ref:  getRef(ctx, valueIntermediate),
			})
		}
	}
	return nil
}

func (chart *GarupaChart) ConvertToSonolus() (resource.LevelData, error) {
	timingGroups := chart.TimingGroupIndexMap()
	if len(timingGroups) > MaxTimingGroups {
		return resource.LevelData{}, tooManyTimingGroupsError(len(timingGroups))
	}
	if svCount := chart.ScrollVelocityCount(); svCount > MaxScrollVelocities {
		return resource.LevelData{}, tooManyScrollVelocitiesError(svCount)
	}

	ctxValues := convertContextValue{
		Entities:             []*resource.LevelDataEntity{},
		BeatToIntermediates:  map[float64][]*Intermediate{},
		IntermediateToRef:    map[*Intermediate]string{},
		IntermediateToEntity: map[*Intermediate]*resource.LevelDataEntity{},
		TimingGroups:         timingGroups,
		refCounter:           0,
	}
	ctx := context.WithValue(context.Background(), contextValuesKey, &ctxValues)
	if err := appendIntermediate(ctx, &Intermediate{Archetype: archetypeInitialization, Data: map[resource.EngineArchetypeDataName]any{}, Sim: false}); err != nil {
		return resource.LevelData{}, err
	}
	if err := appendIntermediate(ctx, &Intermediate{Archetype: archetypeStage, Data: map[resource.EngineArchetypeDataName]any{}, Sim: false}); err != nil {
		return resource.LevelData{}, err
	}

	svs := make([]indexedScrollVelocity, 0, chart.ScrollVelocityCount())
	for objectIndex, object := range *chart {
		if sv, ok := object.(*GarupaSV); ok {
			svs = append(svs, indexedScrollVelocity{
				Beat:        sv.Beat,
				Value:       sv.Value,
				Group:       timingGroupIndex(ctx, sv.TimingGroup),
				ObjectIndex: objectIndex,
			})
			continue
		}
		if err := object.Convert(ctx); err != nil {
			return resource.LevelData{}, err
		}
	}
	sortScrollVelocities(svs)
	for _, sv := range svs {
		if err := appendScrollVelocity(ctx, sv.Beat, sv.Value, sv.Group); err != nil {
			return resource.LevelData{}, err
		}
	}

	beats := make([]float64, 0, len(ctxValues.BeatToIntermediates))
	for beat := range ctxValues.BeatToIntermediates {
		beats = append(beats, beat)
	}
	sort.Float64s(beats)

	for _, beat := range beats {
		intermediates := ctxValues.BeatToIntermediates[beat]
		for i := 1; i < len(intermediates); i++ {
			if err := appendIntermediate(ctx, &Intermediate{
				Archetype: archetypeSimLine,
				Data: map[resource.EngineArchetypeDataName]any{
					dataNameA:     intermediates[i-1],
					dataNameB:     intermediates[i],
					dataNameGroup: simLineGroup(intermediates[i-1], intermediates[i]),
				},
				Sim: false,
			}); err != nil {
				return resource.LevelData{}, err
			}
		}
	}

	entities := make([]resource.LevelDataEntity, len(ctxValues.Entities))
	for i, entity := range ctxValues.Entities {
		entities[i] = *entity
	}

	return resource.LevelData{
		BGMOffset: 0,
		Entities:  entities,
	}, nil
}

type indexedScrollVelocity struct {
	Beat        float64
	Value       float64
	Group       int
	ObjectIndex int
}

func sortScrollVelocities(svs []indexedScrollVelocity) {
	sort.SliceStable(svs, func(i, j int) bool {
		if svs[i].Beat != svs[j].Beat {
			return svs[i].Beat < svs[j].Beat
		}
		if (svs[i].Group == 0) != (svs[j].Group == 0) {
			return svs[i].Group != 0
		}
		return svs[i].ObjectIndex < svs[j].ObjectIndex
	})
}

func simLineGroup(a, b *Intermediate) int {
	aGroup := intermediateGroup(a)
	bGroup := intermediateGroup(b)
	if aGroup == bGroup {
		return aGroup
	}
	return 0
}

func intermediateGroup(intermediate *Intermediate) int {
	group, ok := intermediate.Data[dataNameGroup].(float64)
	if ok {
		return int(group)
	}
	intGroup, ok := intermediate.Data[dataNameGroup].(int)
	if ok {
		return intGroup
	}
	return 0
}
