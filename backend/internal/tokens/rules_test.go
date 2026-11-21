package tokens

import (
	"context"
	"testing"
)

func TestPlacePermissions(t *testing.T) {
	s, repo, n := newSvc(t)
	ctx := context.Background()
	place(t, s, gm, charB, 1, 1)
	place(t, s, ana, charA, 2, 2)
	if len(repo.tokens) != 2 || len(n.events) != 2 || n.events[0] != EventPlaced {
		t.Fatalf("tokens=%d events=%v", len(repo.tokens), n.events)
	}
	wantErr(t, s.Place(ctx, ana, camp, sceneID, charA, 3, 3), "já está na cena")
	wantErr(t, s.Place(ctx, ana, camp, sceneID, charB, 3, 3), "só pode colocar")
	wantErr(t, s.Place(ctx, gm, camp, sceneID, charA, 3, 3), "já está na cena")
	if len(repo.tokens) != 2 {
		t.Fatal("duplicate token created")
	}
}

func TestPlaceRejectsWall(t *testing.T) {
	s, _, _ := newSvc(t)
	wantErr(t, s.Place(context.Background(), gm, camp, sceneID, charA, 5, 3), "bloqueando")
}

func TestPlayerMovesOnlyOwnToken(t *testing.T) {
	s, repo, n := newSvc(t)
	ctx := context.Background()
	place(t, s, gm, charA, 1, 1)
	place(t, s, gm, charB, 2, 1)
	a, b := tokenOf(repo, charA), tokenOf(repo, charB)
	if err := s.Move(ctx, ana, camp, a.ID, 3, 3, false); err != nil {
		t.Fatal(err)
	}
	if a.X != 3 || a.Z != 3 || n.events[len(n.events)-1] != EventMoved {
		t.Fatalf("not moved: %+v %v", a, n.events)
	}
	wantErr(t, s.Move(ctx, ana, camp, b.ID, 3, 3, false), "seus personagens")
	if err := s.Move(ctx, gm, camp, b.ID, 4, 4, false); err != nil {
		t.Fatal(err)
	}
}

func TestMoveBlockedByWallAndStep(t *testing.T) {
	s, repo, _ := newSvc(t)
	ctx := context.Background()
	place(t, s, gm, charA, 3, 3)
	a := tokenOf(repo, charA)
	wantErr(t, s.Move(ctx, ana, camp, a.ID, 5, 3, false), "bloqueando")
	wantErr(t, s.Move(ctx, ana, camp, a.ID, 7, 3, false), "caminho")
	place(t, s, gm, charB, 8, 3)
	b := tokenOf(repo, charB)
	wantErr(t, s.Move(ctx, beto, camp, b.ID, 9, 3, false), "caminho")
	wantErr(t, s.Move(ctx, ana, camp, a.ID, 99, 3, false), "fora do mapa")
	if a.X != 3 {
		t.Fatal("blocked move persisted")
	}
}

func TestGMForceIgnoresBlock(t *testing.T) {
	s, repo, _ := newSvc(t)
	ctx := context.Background()
	place(t, s, gm, charA, 3, 3)
	a := tokenOf(repo, charA)
	wantErr(t, s.Move(ctx, gm, camp, a.ID, 7, 3, false), "caminho")
	if err := s.Move(ctx, gm, camp, a.ID, 7, 3, true); err != nil || a.X != 7 {
		t.Fatalf("force failed: %v %+v", err, a)
	}
	wantErr(t, s.Move(ctx, ana, camp, a.ID, 3, 3, true), "caminho")
}

func TestRemoveGMOnly(t *testing.T) {
	s, repo, _ := newSvc(t)
	ctx := context.Background()
	place(t, s, ana, charA, 1, 1)
	a := tokenOf(repo, charA)
	wantErr(t, s.Remove(ctx, ana, camp, a.ID), "permissão")
	if err := s.Remove(ctx, gm, camp, a.ID); err != nil || len(repo.tokens) != 0 {
		t.Fatalf("remove: %v", err)
	}
}

func TestPingAndFocus(t *testing.T) {
	s, _, n := newSvc(t)
	if err := s.Ping(camp, ana, sceneID, 3.5, 4); err != nil || n.events[0] != EventPinged {
		t.Fatalf("ping: %v %v", err, n.events)
	}
	wantErr(t, s.FocusAll(camp, ana, sceneID, 1, 1), "permissão")
	if err := s.FocusAll(camp, gm, sceneID, 1, 1); err != nil || n.events[1] != EventFocused {
		t.Fatalf("focus: %v", err)
	}
	wantErr(t, s.Ping(camp, ana, "bad", 1, 1), "inválida")
}
