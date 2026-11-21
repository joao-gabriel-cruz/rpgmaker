// Package ambience owns per-scene time of day and weather, token conditions and in-character chat.
package ambience

// Times of day and weathers accepted by ambience.set.
var (
	timesOfDay = []string{"dawn", "day", "dusk", "night"}
	weathers   = []string{"clear", "rain", "storm", "fog", "snow", "ash"}
)

// Ambience is the persisted mood of a scene.
type Ambience struct {
	SceneID   string `json:"sceneId"`
	TimeOfDay string `json:"timeOfDay"`
	Weather   string `json:"weather"`
}

// Default is the ambience of a scene nobody configured.
func Default(sceneID string) Ambience {
	return Ambience{SceneID: sceneID, TimeOfDay: "day", Weather: "clear"}
}

func oneOf(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
