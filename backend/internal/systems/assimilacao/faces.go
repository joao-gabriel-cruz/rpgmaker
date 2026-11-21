package assimilacao

// Die kinds.
const (
	D6  = "d6"
	D10 = "d10"
	D12 = "d12"
)

// Symbols is a tally of rolled symbols.
type Symbols struct {
	Success    int `json:"success"`
	Adaptation int `json:"adaptation"`
	Pressure   int `json:"pressure"`
}

// Add returns the sum of two tallies.
func (a Symbols) Add(b Symbols) Symbols {
	return Symbols{a.Success + b.Success, a.Adaptation + b.Adaptation, a.Pressure + b.Pressure}
}

// faceTable is spec 4.2 (to be validated against the book), indexed by face.
var faceTable = [13]Symbols{
	3: {Pressure: 1}, 4: {Pressure: 1},
	5:  {Adaptation: 1, Pressure: 1},
	6:  {Success: 1},
	7:  {Success: 2},
	8:  {Success: 1, Adaptation: 1},
	9:  {Success: 1, Adaptation: 1, Pressure: 1},
	10: {Success: 2, Pressure: 1},
	11: {Success: 1, Adaptation: 2, Pressure: 1},
	12: {Pressure: 2},
}

// Sides returns the number of faces of a die kind (0 if unknown).
func Sides(kind string) int {
	switch kind {
	case D6:
		return 6
	case D10:
		return 10
	case D12:
		return 12
	}
	return 0
}

// FaceSymbols returns the symbols of a face; zero if the face is out of range for the kind.
func FaceSymbols(kind string, face int) Symbols {
	if face < 1 || face > Sides(kind) {
		return Symbols{}
	}
	return faceTable[face]
}
