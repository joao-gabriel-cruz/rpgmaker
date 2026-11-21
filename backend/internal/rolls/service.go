package rolls

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// ErrAlreadyFinal means the roll was already finalized.
var ErrAlreadyFinal = errors.New("roll already final")

// Repository is the persistence the Service needs.
type Repository interface {
	Insert(ctx context.Context, r Record) (Record, error)
	Get(ctx context.Context, campaignID, id string) (Record, error)
	Finalize(ctx context.Context, campaignID, id string, result json.RawMessage) error
	Recent(ctx context.Context, campaignID string, limit int, includeGM bool) ([]Record, error)
	Usage(ctx context.Context, sessionID, characterID string) (Usage, error)
	Difficulty(ctx context.Context, campaignID, characterID string) (int, error)
	SetDifficulty(ctx context.Context, campaignID, characterID string, d12 int) error
}

// Sessions tells which play session is running ("" when none).
type Sessions interface {
	ActiveID(ctx context.Context, campaignID string) (string, error)
}

// Characters is the part of characters.Service used here.
type Characters interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
	Save(ctx context.Context, m campaigns.Member, id string, sheet json.RawMessage) (characters.View, error)
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the roll rules.
type Service struct {
	repo   Repository
	chars  Characters
	sess   Sessions
	events Events
	rng    assimilacao.RNG
	seed   func() int64
	hooks  []func(ctx context.Context, campaignID, rollID string)
}

// NewService builds a Service using crypto/rand.
func NewService(repo Repository, chars Characters, sess Sessions, ev Events) *Service {
	return &Service{repo: repo, chars: chars, sess: sess, events: ev, rng: assimilacao.CryptoRNG{}, seed: randomSeed}
}

func randomSeed() int64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		panic(err)
	}
	return int64(binary.BigEndian.Uint64(b[:]) >> 1)
}

// RequestInput is the payload of roll.request; DifficultyD12 is honored only for the Mestre.
type RequestInput struct {
	CharacterID string `json:"characterId"`
	Instinct    string `json:"instinct"`
	Aptitude    string `json:"aptitude,omitempty"`
	Secret      bool   `json:"secret,omitempty"`

	SecondInstinct  string `json:"secondInstinct,omitempty"`
	Effort          bool   `json:"effort,omitempty"`
	UseOccupation   bool   `json:"useOccupation,omitempty"`
	UseMarkingEvent bool   `json:"useMarkingEvent,omitempty"`
	DifficultyD12   int    `json:"difficultyD12,omitempty"`

	// BonusKept is set only by trusted callers (actions), never read from a client payload.
	BonusKept int `json:"-"`
}

func member(campaignID string, s realtime.Sender) campaigns.Member {
	return campaigns.Member{CampaignID: campaignID, UserID: s.UserID, Role: s.Role}
}

func visibilityOf(secret bool) string {
	if secret {
		return VisGM
	}
	return VisPublic
}

func filterFor(visibility string) realtime.Filter {
	if visibility == VisGM {
		return realtime.ToGM
	}
	return realtime.ToAll
}

// Request draws an Assimilação roll for a character and stores it as pending.
func (s *Service) Request(ctx context.Context, campaignID string, sender realtime.Sender, in RequestInput) error {
	_, err := s.RequestEntry(ctx, campaignID, sender, in)
	return err
}

// RequestEntry is Request returning the broadcast entry (its RollID links follow-ups such as actions).
func (s *Service) RequestEntry(ctx context.Context, campaignID string, sender realtime.Sender, in RequestInput) (Entry, error) {
	view, sheet, err := s.authorize(ctx, campaignID, sender, in)
	if err != nil {
		return Entry{}, err
	}
	req := requestOf(in)
	sessionID, err := s.prepare(ctx, campaignID, sender, view.ID, &req)
	if err != nil {
		return Entry{}, err
	}
	pool, err := assimilacao.BuildRoll(sheet, req)
	if err != nil {
		return Entry{}, err
	}
	if err := s.commitCosts(ctx, campaignID, sender, view, sheet, req); err != nil {
		return Entry{}, err
	}
	roll := assimilacao.Resolve(pool, s.rng)
	return s.store(ctx, campaignID, sender, Record{
		CharacterID: &view.ID, PlaySessionID: nullID(sessionID), Visibility: visibilityOf(in.Secret),
		Kind: KindAssimilacao, Status: StatusPending,
	}, assimilacaoSpec{Request: req, Pool: pool}, roll.Dice, assimilacao.SuggestKept(roll, pool))
}

// authorize checks who may roll and loads the character with its parsed sheet.
func (s *Service) authorize(ctx context.Context, campaignID string, sender realtime.Sender, in RequestInput) (characters.View, assimilacao.Sheet, error) {
	if in.Secret && sender.Role != campaigns.RoleGM {
		return characters.View{}, assimilacao.Sheet{}, realtime.UserError("Apenas o Mestre pode rolar em segredo.")
	}
	view, err := s.chars.Get(ctx, member(campaignID, sender), in.CharacterID)
	if err != nil {
		return view, assimilacao.Sheet{}, err
	}
	if sender.Role != campaigns.RoleGM && (view.OwnerID == nil || *view.OwnerID != sender.UserID) {
		return view, assimilacao.Sheet{}, realtime.UserError("Você só pode rolar pelos seus personagens.")
	}
	sheet, err := assimilacao.ParseSheet(view.Sheet)
	if err != nil {
		return view, sheet, realtime.UserError("A ficha deste personagem é inválida.")
	}
	return view, sheet, nil
}

func (s *Service) store(ctx context.Context, campaignID string, sender realtime.Sender, rec Record, spec any, dice []assimilacao.Die, suggested []int) (Entry, error) {
	rec.CampaignID, rec.UserID, rec.Seed = campaignID, sender.UserID, s.seed()
	rec.Spec, _ = json.Marshal(spec)
	rec.Dice, _ = json.Marshal(dice)
	saved, err := s.repo.Insert(ctx, rec)
	if err != nil {
		return Entry{}, err
	}
	entry, err := toEntry(saved)
	if err != nil {
		return Entry{}, err
	}
	entry.Suggested = suggested
	s.events.Broadcast(campaignID, filterFor(saved.Visibility), EventRolled, entry)
	return entry, nil
}

// OnFinalized registers a callback run synchronously after an Assimilação roll is finalized.
func (s *Service) OnFinalized(fn func(ctx context.Context, campaignID, rollID string)) {
	s.hooks = append(s.hooks, fn)
}

func (s *Service) runHooks(ctx context.Context, campaignID, rollID string) {
	for _, h := range s.hooks {
		h(ctx, campaignID, rollID)
	}
}

// Keep finalizes a pending roll with the dice chosen by whoever rolled it.
func (s *Service) Keep(ctx context.Context, campaignID string, sender realtime.Sender, rollID string, kept []int) error {
	rec, err := s.repo.Get(ctx, campaignID, rollID)
	if err != nil {
		return err
	}
	if rec.UserID != sender.UserID {
		return realtime.UserError("Apenas quem rolou pode escolher os dados.")
	}
	if rec.Kind != KindAssimilacao || rec.Status != StatusPending {
		return ErrAlreadyFinal
	}
	res, err := s.tally(ctx, member(campaignID, sender), rec, kept)
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(res)
	if err := s.repo.Finalize(ctx, campaignID, rollID, raw); err != nil {
		return err
	}
	s.events.Broadcast(campaignID, filterFor(rec.Visibility), EventFinalized, map[string]any{"rollId": rollID, "result": res})
	s.runHooks(ctx, campaignID, rollID)
	return nil
}

// tally settles the kept dice into the final result, applying the character's current health penalty.
func (s *Service) tally(ctx context.Context, m campaigns.Member, rec Record, kept []int) (finalResult, error) {
	spec, dice, err := decodeRoll(rec)
	if err != nil {
		return finalResult{}, err
	}
	derived := s.derivedFor(ctx, m, rec.CharacterID)
	base, err := assimilacao.Finalize(assimilacao.Roll{Dice: dice}, spec.Pool, kept, derived)
	return finalResult{Result: base, HealthLevel: derived.LevelName}, err
}

func decodeRoll(rec Record) (assimilacaoSpec, []assimilacao.Die, error) {
	var spec assimilacaoSpec
	var dice []assimilacao.Die
	if json.Unmarshal(rec.Spec, &spec) != nil || json.Unmarshal(rec.Dice, &dice) != nil {
		return spec, nil, errors.New("corrupt roll")
	}
	return spec, dice, nil
}

// derivedFor reads the current health penalty; a deleted character has none.
func (s *Service) derivedFor(ctx context.Context, m campaigns.Member, characterID *string) assimilacao.Derived {
	if characterID == nil {
		return assimilacao.Derived{}
	}
	view, err := s.chars.Get(ctx, m, *characterID)
	if err != nil {
		return assimilacao.Derived{}
	}
	sheet, err := assimilacao.ParseSheet(view.Sheet)
	if err != nil {
		return assimilacao.Derived{}
	}
	return assimilacao.Derive(sheet)
}

// Generic rolls an XdY+Z expression and finalizes it immediately.
func (s *Service) Generic(ctx context.Context, campaignID string, sender realtime.Sender, expression string) error {
	spec, err := parseExpression(expression)
	if err != nil {
		return err
	}
	kind := "d" + itoa(spec.Sides)
	dice := make([]assimilacao.Die, spec.Count)
	total := spec.Modifier
	for i := range dice {
		dice[i] = assimilacao.Die{Kind: kind, Face: s.rng.Intn(spec.Sides) + 1}
		total += dice[i].Face
	}
	result, _ := json.Marshal(map[string]int{"total": total})
	_, err = s.store(ctx, campaignID, sender, Record{Visibility: VisPublic, Kind: KindGeneric, Status: StatusFinal, Result: result}, spec, dice, nil)
	return err
}

// Recent lists the latest rolls the member may see, newest first.
func (s *Service) Recent(ctx context.Context, m campaigns.Member, limit int) ([]Entry, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	recs, err := s.repo.Recent(ctx, m.CampaignID, limit, m.Role == campaigns.RoleGM)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(recs))
	for _, r := range recs {
		e, err := toEntry(r)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
