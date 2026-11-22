package refuge

import (
	"context"
	"testing"
)

func TestPreviewDoesNotWriteAndConfirmDoes(t *testing.T) {
	s, repo, _, ev := setup(sheet)
	repo.r = weekRefuge(done(KindQuarry, 2))
	var hooked []WeekReport
	s.OnWeekAdvanced(func(_ context.Context, id string, rep WeekReport) {
		if id != "c" {
			t.Errorf("campaign = %s", id)
		}
		hooked = append(hooked, rep)
	})
	rep, err := s.PreviewWeek(ctx, gm, nil)
	if err != nil || repo.r.Week != 1 || len(repo.r.History) != 0 || len(ev.types) != 0 || len(hooked) != 0 || rep.Entered["minerals"] != 4 {
		t.Fatalf("preview = %+v %v week %d", rep, err, repo.r.Week)
	}
	if _, err := s.PreviewWeek(ctx, ana, nil); err == nil {
		t.Error("player previewed")
	}
	rep, err = s.ConfirmWeek(ctx, gm, nil)
	if err != nil || repo.r.Week != 2 || len(repo.r.History) != 1 || len(hooked) != 1 || hooked[0].Week != 1 || repo.r.Resources["minerals"] != 4 {
		t.Fatalf("confirm = %+v %v", rep, err)
	}
	if len(ev.types) != 2 || ev.types[0] != EventUpdated || ev.types[1] != EventWeekAdvanced {
		t.Errorf("events = %v", ev.types)
	}
	if _, err := s.ConfirmWeek(ctx, ana, nil); err == nil {
		t.Error("player confirmed")
	}
}

func TestHistoryKeepsLast20Weeks(t *testing.T) {
	s, repo, _, _ := setup(sheet)
	repo.r = weekRefuge()
	repo.r.Resources["water"], repo.r.Resources["food"], repo.r.Resources["wood"] = 1000, 1000, 1000
	repo.r.ReservesMax = 5000
	for range 25 {
		if _, err := s.ConfirmWeek(ctx, gm, nil); err != nil {
			t.Fatal(err)
		}
	}
	h := repo.r.History
	if len(h) != historyKeep || h[0].Week != 6 || h[len(h)-1].Week != 25 {
		t.Errorf("history = %d first %d", len(h), h[0].Week)
	}
}
