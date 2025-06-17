package inmemory

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/AnruKitakaze/tower-of-hanoi/internal/domain"
)

type playerInmemoryRepo struct {
	users        map[domain.PlayerID]domain.Player
	nameToID     map[string]domain.PlayerID
	lock         sync.RWMutex
	nameToIDLock sync.RWMutex
	logger       *slog.Logger
}

func NewPlayerInmemoryRepo(logger *slog.Logger) *playerInmemoryRepo {
	return &playerInmemoryRepo{
		users:        make(map[domain.PlayerID]domain.Player),
		nameToID:     make(map[string]domain.PlayerID),
		lock:         sync.RWMutex{},
		nameToIDLock: sync.RWMutex{},
		logger:       logger,
	}
}

func (r *playerInmemoryRepo) Save(ctx context.Context, nickname string) (domain.PlayerID, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.nameToIDLock.Lock()
	defer r.nameToIDLock.Unlock()

	id := domain.PlayerID(len(r.users) + 1)
	// TODO: generate ID
	r.users[id] = domain.Player{ID: id, Nickname: nickname}
	r.nameToID[nickname] = id

	r.logger.Info("player successfully created",
		slog.Int("id", int(id)),
		slog.String("nickname", nickname),
	)

	return id, nil
}

func (r *playerInmemoryRepo) GetByID(ctx context.Context, id domain.PlayerID) (*domain.Player, error) {
	r.lock.RLock()
	p, ok := r.users[id]
	r.lock.RUnlock()

	if !ok {
		return nil, domain.ErrPlayerNotFound
	}

	return &domain.Player{ID: p.ID, Nickname: p.Nickname}, nil
}

func (r *playerInmemoryRepo) GetAll(ctx context.Context) ([]*domain.Player, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	players := []*domain.Player{}

	for _, v := range r.users {
		players = append(players, &domain.Player{ID: v.ID, Nickname: v.Nickname})
	}

	return players, nil
}

func (r *playerInmemoryRepo) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	r.nameToIDLock.RLock()
	defer r.nameToIDLock.RUnlock()

	id, ok := r.nameToID[nickname]
	if !ok {
		return nil, domain.ErrPlayerNotFound
	}
	p, ok := r.users[id]

	if !ok {
		return nil, fmt.Errorf("cannot get player: integrity error")
	}

	return &domain.Player{ID: p.ID, Nickname: p.Nickname}, nil
}
