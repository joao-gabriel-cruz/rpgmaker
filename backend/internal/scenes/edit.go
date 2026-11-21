package scenes

import (
	"context"
	"fmt"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// audience: players only ever receive events of the active scene; the origin already applied the change locally.
func audience(sc Scene, origin string) realtime.Filter {
	return func(s realtime.Sender) bool {
		if origin != "" && s.UserID == origin {
			return false
		}
		return sc.IsActive || realtime.ToGM(s)
	}
}

func (s *Service) load(ctx context.Context, campaignID, sceneID string) (Scene, error) {
	if !uuidRe.MatchString(sceneID) {
		return Scene{}, ErrNotFound
	}
	return s.repo.Get(ctx, campaignID, sceneID)
}

// Paint validates and persists a terrain patch, then broadcasts it.
func (s *Service) Paint(ctx context.Context, campaignID, sceneID, origin string, patch maps.Patch) error {
	l := s.lockFor(sceneID)
	l.Lock()
	defer l.Unlock()
	sc, err := s.load(ctx, campaignID, sceneID)
	if err != nil {
		return err
	}
	g, err := maps.Deserialize(working(sc).Grid)
	if err != nil {
		return err
	}
	if len(patch.Cells) == 0 || len(patch.Cells) > len(g.Cells)*2 {
		return fmt.Errorf("%w: invalid cell count", ErrInvalidPatch)
	}
	if err := g.ApplyPatch(patch); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidPatch, err.Error())
	}
	if err := s.saveGrid(ctx, sc, g.Serialize().Data); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, editAudience(sc, origin), eventName(sc, EventPatched, EventDraftPatched), map[string]any{"sceneId": sceneID, "patch": patch})
	return nil
}

// EditObjects validates and persists an object patch, then broadcasts it.
func (s *Service) EditObjects(ctx context.Context, campaignID, sceneID, origin string, patch ObjectPatch) error {
	l := s.lockFor(sceneID)
	l.Lock()
	defer l.Unlock()
	sc, err := s.load(ctx, campaignID, sceneID)
	if err != nil {
		return err
	}
	next, err := applyObjects(working(sc), patch)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidPatch, err.Error())
	}
	if err := s.saveObjects(ctx, sc, next); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, editAudience(sc, origin), eventName(sc, EventObjectsPatched, EventDraftObjectsPatched), map[string]any{"sceneId": sceneID, "patch": normalize(patch)})
	return nil
}

func normalize(p ObjectPatch) ObjectPatch {
	if p.Add == nil {
		p.Add = []maps.Object{}
	}
	if p.Remove == nil {
		p.Remove = []string{}
	}
	if p.Update == nil {
		p.Update = []maps.Object{}
	}
	return p
}

// applyObjects validates the whole patch and returns the resulting list (update, then remove, then add).
func applyObjects(sc Scene, p ObjectPatch) ([]maps.Object, error) {
	g := &maps.Grid{Width: sc.Width, Depth: sc.Depth}
	out, err := updateObjects(g, sc.Objects, p)
	if err != nil {
		return nil, err
	}
	removed := map[string]bool{}
	for _, id := range p.Remove {
		removed[id] = true
	}
	kept := out[:0:0]
	ids := map[string]bool{}
	for _, o := range out {
		if !removed[o.ID] {
			kept = append(kept, o)
			ids[o.ID] = true
		}
	}
	return addObjects(g, kept, ids, p.Add)
}

func addObjects(g *maps.Grid, kept []maps.Object, ids map[string]bool, add []maps.Object) ([]maps.Object, error) {
	for _, o := range add {
		if err := checkObject(g, o); err != nil {
			return nil, err
		}
		if ids[o.ID] {
			return nil, fmt.Errorf("objeto duplicado: %q", o.ID)
		}
		ids[o.ID] = true
		kept = append(kept, o)
	}
	if len(kept) > maxObjects {
		return nil, fmt.Errorf("limite de objetos excedido")
	}
	return kept, nil
}

func updateObjects(g *maps.Grid, cur []maps.Object, p ObjectPatch) ([]maps.Object, error) {
	out := append([]maps.Object{}, cur...)
	idx := map[string]int{}
	for i, o := range out {
		idx[o.ID] = i
	}
	for _, o := range p.Update {
		i, ok := idx[o.ID]
		if !ok {
			return nil, fmt.Errorf("objeto inexistente: %q", o.ID)
		}
		if err := checkObject(g, o); err != nil {
			return nil, err
		}
		out[i] = o
	}
	return out, nil
}

func checkObject(g *maps.Grid, o maps.Object) error {
	if o.ID == "" || len(o.ID) > 64 {
		return fmt.Errorf("id de objeto inválido")
	}
	if !maps.KnownPrefab(o.Prefab) {
		return fmt.Errorf("objeto desconhecido: %q", o.Prefab)
	}
	if o.Rotation != 0 && o.Rotation != 90 && o.Rotation != 180 && o.Rotation != 270 {
		return fmt.Errorf("rotação inválida: %d", o.Rotation)
	}
	for _, c := range maps.Footprint(o) {
		if !g.InBounds(c.X, c.Z) {
			return fmt.Errorf("objeto fora do mapa: (%d, %d)", c.X, c.Z)
		}
	}
	return nil
}
