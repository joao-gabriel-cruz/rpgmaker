package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Card is an Assimilação drawn by a rest test, waiting for the Mestre to confirm it.
type Card struct {
	Suit        string `json:"suit"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var suitLabel = map[string]string{"hearts": "♥ Evolutiva", "diamonds": "♦ Adaptativa", "spades": "♠ Inoportuna"}

// deck holds invented example cards; Mestres are expected to confirm or reword them at the table.
var deck = map[string][]Card{
	"hearts": {
		{"hearts", "Pele Calejada", "O corpo se acostuma ao desconforto e aguenta um pouco mais."},
		{"hearts", "Olhar Atento", "Detalhes que passavam batido agora saltam aos olhos."},
		{"hearts", "Fôlego Renovado", "Os pulmões se adaptam ao ar pesado das ruínas."},
	},
	"diamonds": {
		{"diamonds", "Passos Leves", "Os pés aprendem a pisar onde o chão não range."},
		{"diamonds", "Mãos Hábeis", "Os dedos encontram soluções que a cabeça ainda não pensou."},
		{"diamonds", "Instinto de Abrigo", "Sempre sabe onde há um canto seguro para se esconder."},
	},
	"spades": {
		{"spades", "Sono Agitado", "Sonhos estranhos tiram o sossego nas noites seguintes."},
		{"spades", "Tremor nas Mãos", "Quando a tensão aperta, os dedos não obedecem direito."},
		{"spades", "Voz Rouca", "Algo mudou no timbre, e nem todo mundo gosta de ouvir."},
	},
}

// drawCards draws one card per symbol type that showed up: Success -> hearts, Adaptation -> diamonds, Pressure -> spades.
func (s *Service) drawCards(sym assimilacao.Symbols) []Card {
	var out []Card
	for _, p := range []struct {
		suit string
		n    int
	}{{"hearts", sym.Success}, {"diamonds", sym.Adaptation}, {"spades", sym.Pressure}} {
		if p.n > 0 {
			out = append(out, deck[p.suit][s.pick(len(deck[p.suit]))])
		}
	}
	return out
}

// rest is the Descanso test: suggests regaining Determinação (more in a safe place) and draws cards for the Mestre.
func (s *Service) rest(ctx context.Context, campaignID string, r *run, sym assimilacao.Symbols) (outcome, error) {
	regain := 0
	if sym.Success > 0 {
		regain = 1
		if r.safe {
			regain = 2
		}
	}
	cards := s.drawCards(sym)
	msg := restMessage(r, regain, cards)
	if err := s.suggest(ctx, r, r.char.ID, r.char.Name, effects.EffectNotice, 0, msg); err != nil {
		return outcome{}, err
	}
	if len(cards) > 0 {
		s.events.Broadcast(campaignID, realtime.ToGM, EventCards, map[string]any{
			"actionRunId": r.id, "characterId": r.char.ID, "characterName": r.char.Name, "cards": cards})
	}
	text := r.char.Name + " descansou"
	if r.safe {
		text += " em um local seguro"
	}
	return outcome{KindRest, text + "."}, nil
}

func restMessage(r *run, regain int, cards []Card) string {
	place := ""
	if r.safe {
		place = " (local seguro)"
	}
	msg := fmt.Sprintf("Descanso de %s%s: ", r.char.Name, place)
	if regain > 0 {
		msg += fmt.Sprintf("recuperar %d de Determinação.", regain)
	} else {
		msg += "sem Sucessos, nenhuma Determinação recuperada."
	}
	if len(cards) == 0 {
		return msg
	}
	names := make([]string, len(cards))
	for i, c := range cards {
		names[i] = suitLabel[c.Suit] + " " + c.Name
	}
	return msg + " Cartas: " + strings.Join(names, ", ") + "."
}
