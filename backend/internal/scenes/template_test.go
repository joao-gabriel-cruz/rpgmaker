package scenes

import (
	"context"
	"errors"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maptemplates"
)

func TestCreateFromTemplateStoresMapAndAppliesAmbience(t *testing.T) {
	svc, repo, n := newSvc()
	var got [4]string
	svc.OnTemplateAmbience(func(_ context.Context, campaignID, sceneID, tod, weather string) error {
		got = [4]string{campaignID, sceneID, tod, weather}
		return nil
	})
	sc, err := svc.CreateFromTemplate(ctx, camp, " Vila ", "vila-abandonada", 42)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Name != "Vila" || sc.Width != 64 || sc.Depth != 64 || len(sc.Objects) == 0 || sc.IsActive {
		t.Fatalf("scene: %+v", sc)
	}
	if got != [4]string{camp, sc.ID, "dusk", "clear"} {
		t.Errorf("ambience hook: %v", got)
	}
	stored, _ := repo.Get(ctx, camp, sc.ID)
	tpl, _ := maptemplates.Get("vila-abandonada")
	g, objs := tpl.Generate(42)
	if stored.Grid.Data != g.Serialize().Data || len(stored.Objects) != len(objs) {
		t.Error("stored map differs from the generated one")
	}
	if _, ok := n.last(EventListChanged); !ok {
		t.Error("scene list change not announced")
	}
}

func TestCreateFromTemplateRejections(t *testing.T) {
	svc, _, _ := newSvc()
	if _, err := svc.CreateFromTemplate(ctx, camp, "x", "nope", 1); !errors.Is(err, ErrUnknownTemplate) {
		t.Errorf("unknown template: %v", err)
	}
	if _, err := svc.CreateFromTemplate(ctx, camp, " ", "vila-abandonada", 1); !errors.Is(err, ErrInvalidName) {
		t.Errorf("name: %v", err)
	}
}

func TestCreateFromTemplateSurvivesAmbienceFailure(t *testing.T) {
	svc, _, _ := newSvc()
	svc.OnTemplateAmbience(func(context.Context, string, string, string, string) error { return errors.New("boom") })
	if _, err := svc.CreateFromTemplate(ctx, camp, "Rodovia", "rodovia-engarrafada", 1); err != nil {
		t.Fatal(err)
	}
}

func TestLargestSceneIsAccepted(t *testing.T) {
	svc, _, _ := newSvc()
	sc, err := svc.Create(ctx, camp, "Grande", 128, 128)
	if err != nil || len(sc.Grid.Data) != 128*128*2 {
		t.Fatalf("128x128: %v", err)
	}
	g, err := maps.Deserialize(sc.Grid)
	if err != nil || len(g.Cells) != 128*128 {
		t.Fatalf("deserialize: %v", err)
	}
}

func TestStoreAcceptsLargeScenesAndTemplates(t *testing.T) {
	s, campID := testStore(t)
	svc := NewService(s, &fakeNotifier{})
	big, err := svc.Create(ctx, campID, "128", 128, 128)
	if err != nil || big.Width != 128 {
		t.Fatalf("128 scene: %+v %v", big, err)
	}
	sc, err := svc.CreateFromTemplate(ctx, campID, "Rodovia", "rodovia-engarrafada", 9)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, campID, sc.ID)
	if err != nil || got.Width != 96 || got.Depth != 64 || len(got.Objects) != len(sc.Objects) || got.Grid.Data != sc.Grid.Data {
		t.Fatalf("stored template scene: %v", err)
	}
}
