package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"
	"unicode/utf8"

	"github.com/AnruKitakaze/tower-of-hanoi/internal/domain"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// TODO: Move to env later
const DB_PATH = "file://internal/infrastructure/persistance/postgresql/migrations"
const DSN = "postgres://userexample:passwordexample@localhost/your_db"
const DriverName = "pgx"

func RunMigrations(logger *slog.Logger, db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error("could not create migrate driver", slog.Any("err", err))
		return fmt.Errorf("RunMigrations: could not create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		DB_PATH,
		"postgres",
		driver,
	)
	if err != nil {
		logger.Error("failed to setup migration", slog.Any("err", err))
		return fmt.Errorf("RunMigrations: failed to setup migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to apply migration", slog.Any("err", err))
		return fmt.Errorf("RunMigrations: failed to apply migration: %w", err)
	}

	logger.Info("Migrations ran successfully.")
	return nil
}

type playerPostgresRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPlayerPostgresRepo(logger *slog.Logger, db *sql.DB) (*playerPostgresRepo, error) {
	// FIX: To config
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		<-appSignal
		stop()
	}()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		logger.Error("unable to conntect to database", slog.Any("err", err))
		return nil, fmt.Errorf("NewPlayerPostgresRepo: unable to conntect to database: %w", err)
	}
	logger.Debug("database ping ok")

	return &playerPostgresRepo{db: db, logger: logger}, nil
}

func (r *playerPostgresRepo) Save(ctx context.Context, nickname string) (domain.PlayerID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if utf8.RuneCountInString(nickname) > 100 {
		r.logger.Error("will not save player to db: name is too long", slog.String("nickname", nickname))
		return 0, &domain.ErrCannotCreatePlayer{Nickname: nickname, Reason: "name is too long"}
	}

	var id int
	err := r.db.QueryRowContext(ctx, "INSERT INTO users (username) VALUES ($1) RETURNING id", nickname).Scan(&id)
	if err != nil {
		r.logger.Error("failed to save player to db", slog.Any("err", err))
		return 0, &domain.ErrCannotCreatePlayer{Nickname: nickname, Reason: err.Error()}
	}

	return domain.PlayerID(id), nil
}

func (r *playerPostgresRepo) GetByID(ctx context.Context, id domain.PlayerID) (*domain.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var p domain.Player
	err := r.db.QueryRowContext(ctx, "SELECT id, username FROM users WHERE id = $1", id).Scan(&p.ID, &p.Nickname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug("no players found", slog.Int("id", int(id)))
			return nil, domain.ErrPlayerNotFound
		}
		r.logger.Error("failed to get players from db", slog.Any("err", err))
		return nil, fmt.Errorf("GetByID: cannot find user id=%v: %w", id, err)
	}

	return &p, nil
}

func (r *playerPostgresRepo) GetAll(ctx context.Context) ([]*domain.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, "SELECT id, username FROM users")
	if err != nil {
		r.logger.Error("failed to get users", slog.Any("err", err))
		return nil, fmt.Errorf("GetAll: cannot get users: %w", err)
	}

	players := make([]*domain.Player, 0)
	for rows.Next() {
		var p domain.Player
		if err := rows.Scan(&p.ID, &p.Nickname); err != nil {
			r.logger.Error("failed to parse users", slog.Any("err", err))
			return nil, fmt.Errorf("GetAll: cannot parse users: %w", err)
		}
		players = append(players, &p)
	}

	return players, nil
}

func (r *playerPostgresRepo) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var p domain.Player
	err := r.db.QueryRowContext(ctx, "SELECT id, username FROM users WHERE username = $1", nickname).Scan(&p.ID, &p.Nickname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug("no players found", slog.String("nickname", nickname))
			return nil, domain.ErrPlayerNotFound
		}
		r.logger.Error("failed to get players from db", slog.Any("err", err))
		return nil, fmt.Errorf("GetByNickname: cannot find user nickname=%v: %w", nickname, err)
	}

	return &p, nil
}
