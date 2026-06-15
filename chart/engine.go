package chart

const (
	EngineNotGarupa         = "notgarupa"
	EngineNotGarupaHabahiro = "notgarupa-habahiro"
)

func (chart GarupaChart) PreferredEngine() string {
	if chart.requiresHabahiro() {
		return EngineNotGarupaHabahiro
	}
	return EngineNotGarupa
}

func (chart GarupaChart) requiresHabahiro() bool {
	for _, object := range chart {
		switch object := object.(type) {
		case *GarupaNote:
			if noteRequiresHabahiro(object.BaseGarupaNote) {
				return true
			}
		case *GarupaDirectionalNote:
			if noteRequiresHabahiro(object.BaseGarupaNote) {
				return true
			}
		case *GarupaSlide:
			for _, connection := range object.Connections {
				if noteRequiresHabahiro(connection.BaseGarupaNote) {
					return true
				}
			}
		}
	}
	return false
}

func noteRequiresHabahiro(note BaseGarupaNote) bool {
	if note.Type == "Directional" {
		return false
	}
	return normalizedWidth(note.Width) > 1
}
