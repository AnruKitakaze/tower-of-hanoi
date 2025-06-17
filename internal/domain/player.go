package domain

import (
	"context"
	"errors"
	"fmt"
)

var ErrPlayerNotFound = errors.New("player is not found")

type ErrCannotCreatePlayer struct {
	Nickname string
	Reason   string
}

func (e *ErrCannotCreatePlayer) Error() string {
	return fmt.Sprintf("cannot create player '%s': %s", e.Nickname, e.Reason)
}

type PlayerID int

type Player struct {
	ID       PlayerID
	Nickname string
}

type PlayerRepository interface {
	Save(ctx context.Context, nickname string) (PlayerID, error)
	GetByID(ctx context.Context, id PlayerID) (*Player, error)
	GetAll(ctx context.Context) ([]*Player, error)
	GetByNickname(ctx context.Context, nickname string) (*Player, error)
}
