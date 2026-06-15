package chart

type TimingGroupIndexMap map[TimingGroupID]int

func (chart GarupaChart) TimingGroupIndexMap() TimingGroupIndexMap {
	groups := TimingGroupIndexMap{GlobalTimingGroupID: 0}
	nextIndex := 1
	add := func(id TimingGroupID) {
		id = normalizeTimingGroup(id)
		if _, ok := groups[id]; ok {
			return
		}
		groups[id] = nextIndex
		nextIndex++
	}

	for _, object := range chart {
		switch object := object.(type) {
		case *GarupaNote:
			add(object.TimingGroup)
		case *GarupaDirectionalNote:
			add(object.TimingGroup)
		case *GarupaSlide:
			add(object.TimingGroup)
		case *GarupaSV:
			add(object.TimingGroup)
		}
	}
	return groups
}

func (groups TimingGroupIndexMap) Index(id TimingGroupID) int {
	index, ok := groups[normalizeTimingGroup(id)]
	if ok {
		return index
	}
	return groups[GlobalTimingGroupID]
}

func (chart GarupaChart) ScrollVelocityCount() int {
	count := 0
	for _, object := range chart {
		if _, ok := object.(*GarupaSV); ok {
			count++
		}
	}
	return count
}
