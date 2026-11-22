package refuge

import "encoding/json"

// stateJSON is the refuges.state jsonb document: the Sheet plus the lists reserved for later deliveries.
type stateJSON struct {
	Sheet
	Buildings []json.RawMessage `json:"buildings"`
	Crises    []json.RawMessage `json:"crises"`
	Residents []json.RawMessage `json:"residents"`
}

func emptyList(l []json.RawMessage) []json.RawMessage {
	if l == nil {
		return []json.RawMessage{}
	}
	return l
}

func decodeState(raw []byte, r *Refuge) error {
	st := stateJSON{Sheet: DefaultSheet()}
	if err := json.Unmarshal(raw, &st); err != nil {
		return err
	}
	st.normalize()
	r.Sheet = st.Sheet
	r.Buildings, r.Crises, r.Residents = emptyList(st.Buildings), emptyList(st.Crises), emptyList(st.Residents)
	return nil
}

func encodeState(r Refuge) ([]byte, error) {
	return json.Marshal(stateJSON{Sheet: r.Sheet, Buildings: emptyList(r.Buildings), Crises: emptyList(r.Crises), Residents: emptyList(r.Residents)})
}
