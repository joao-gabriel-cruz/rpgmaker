package refuge

import "encoding/json"

// stateJSON is the refuges.state jsonb document: the Sheet plus the lists of later deliveries.
type stateJSON struct {
	Sheet
	Buildings      []Building     `json:"buildings"`
	Proposals      []Proposal     `json:"proposals"`
	History        []WeekReport   `json:"history"`
	Crises         []Crisis       `json:"crises"`
	Residents      []Resident     `json:"residents"`
	Expeditions    []Expedition   `json:"expeditions"`
	CrisisSettings CrisisSettings `json:"crisisSettings"`
}

func emptyList[T any](l []T) []T {
	if l == nil {
		return []T{}
	}
	return l
}

func nonNil[T any](l []T) []T {
	if l == nil {
		return []T{}
	}
	return l
}

func decodeState(raw []byte, r *Refuge) error {
	st := stateJSON{Sheet: DefaultSheet(), CrisisSettings: DefaultCrisisSettings()}
	if err := json.Unmarshal(raw, &st); err != nil {
		return err
	}
	st.normalize()
	r.Sheet = st.Sheet
	r.Crises, r.Residents = emptyList(st.Crises), emptyList(st.Residents)
	r.Buildings, r.Proposals, r.History = nonNil(st.Buildings), nonNil(st.Proposals), nonNil(st.History)
	r.Expeditions, r.CrisisSettings = emptyList(st.Expeditions), st.CrisisSettings
	return nil
}

func encodeState(r Refuge) ([]byte, error) {
	return json.Marshal(stateJSON{Sheet: r.Sheet, Buildings: nonNil(r.Buildings), Proposals: nonNil(r.Proposals), History: nonNil(r.History),
		Crises: emptyList(r.Crises), Residents: emptyList(r.Residents), Expeditions: emptyList(r.Expeditions), CrisisSettings: r.CrisisSettings})
}
