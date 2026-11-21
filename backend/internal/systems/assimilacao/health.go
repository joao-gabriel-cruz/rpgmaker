package assimilacao

var levelNames = [6]string{"Saudável", "Escoriação", "Laceração", "Ferimentos", "Debilitação", "Incapacitação"}

var levelPenalty = [6]int{0, 0, 1, 1, 2, 2}

// Derived holds values computed from a sheet; never stored.
type Derived struct {
	BoxesPerLevel  int    `json:"boxesPerLevel"`
	CurrentLevel   int    `json:"currentLevel"`
	LevelName      string `json:"levelName"`
	SuccessPenalty int    `json:"successPenalty"`
	Debilitated    bool   `json:"debilitated"`
	Incapacitated  bool   `json:"incapacitated"`
	XPSpent        int    `json:"xpSpent"`
	XPAvailable    int    `json:"xpAvailable"`
}

// Derive computes health levels, penalties and XP balance.
func Derive(s Sheet) Derived {
	boxes := 1 + s.Instincts.Potency + s.Instincts.Resolution
	if boxes < 1 {
		boxes = 1
	}
	level := max(s.Health.Damage, 0) / boxes
	level = min(level, len(levelNames)-1)
	spent := 0
	for _, c := range s.Characteristics {
		spent += 2 * c.Level
	}
	return Derived{
		BoxesPerLevel:  boxes,
		CurrentLevel:   level,
		LevelName:      levelNames[level],
		SuccessPenalty: levelPenalty[level],
		Debilitated:    level >= 4,
		Incapacitated:  level >= 5,
		XPSpent:        spent,
		XPAvailable:    s.XP.Total - spent,
	}
}
