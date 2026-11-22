package mcp

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// userError is a PT-BR message that is safe to show Claude and the Mestre as is.
type userError string

func (e userError) Error() string { return string(e) }

func newUserError(msg string) error { return userError(msg) }

// toolError is the single exit for tool errors: known user errors keep their PT-BR text,
// anything else is logged and replaced by a generic message so internals never leak.
func toolError(err error) error {
	if err == nil {
		return nil
	}
	var ue userError
	switch {
	case errors.As(err, &ue):
		return ue
	case errors.Is(err, errNoActiveScene), errors.Is(err, errOpenDraft):
		return err
	case errors.Is(err, characters.ErrInvalidInput):
		return newUserError("Dados inválidos para o personagem (nome longo demais?)")
	case errors.Is(err, scenes.ErrNotFound):
		return newUserError("Cena não encontrada")
	case errors.Is(err, scenes.ErrInvalidPatch):
		return newUserError("Mudança inválida no mapa: " + strings.TrimPrefix(err.Error(), "scenes: invalid patch: "))
	}
	if msg, ok := realtime.UserMessage(err); ok {
		return newUserError(msg)
	}
	slog.Error("mcp tool", "err", err)
	return newUserError("Erro interno. Tente novamente")
}
