package scenes

import (
	"context"
	"log/slog"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maptemplates"
)

// AmbienceSetter applies the suggested time of day and weather to a new scene (wired to internal/ambience).
type AmbienceSetter func(ctx context.Context, campaignID, sceneID, timeOfDay, weather string) error

// OnTemplateAmbience registers the hook that applies a template's suggested ambience; call it before serving.
func (s *Service) OnTemplateAmbience(fn AmbienceSetter) { s.ambience = fn }

// CreateFromTemplate generates the template for the seed and stores it as a new inactive scene.
func (s *Service) CreateFromTemplate(ctx context.Context, campaignID, name, templateID string, seed int64) (Scene, error) {
	name, err := cleanName(name)
	if err != nil {
		return Scene{}, err
	}
	tpl, ok := maptemplates.Get(templateID)
	if !ok {
		return Scene{}, ErrUnknownTemplate
	}
	built := tpl.Build(seed)
	sc, err := s.repo.Create(ctx, campaignID, name, built.Grid.Width, built.Grid.Depth, built.Grid.Data)
	if err != nil {
		return Scene{}, err
	}
	if err := s.repo.Replace(ctx, sc.ID, built.Grid.Data, built.Objects); err != nil {
		_ = s.repo.Delete(ctx, campaignID, sc.ID)
		return Scene{}, err
	}
	sc.Grid, sc.Objects = built.Grid, built.Objects
	if s.ambience != nil {
		if err := s.ambience(ctx, campaignID, sc.ID, built.Ambience.TimeOfDay, built.Ambience.Weather); err != nil {
			slog.Warn("template ambience not applied", "scene", sc.ID, "err", err)
		}
	}
	s.changed(campaignID)
	return sc, nil
}
