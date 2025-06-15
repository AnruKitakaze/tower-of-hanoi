package domain

import (
	"context"
	"time"
)

type Player struct {
	ID       int
	Nickname string
}

type UserRepository interface {
	Save(ctx context.Context, user User) error
	GetByID(ctx context.Context, id UserID) (User, error)
}

type UserID int

type User struct {
	ID        UserID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
