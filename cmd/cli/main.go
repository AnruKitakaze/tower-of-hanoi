package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AnruKitakaze/tower-of-hanoi/internal/domain"
	"github.com/AnruKitakaze/tower-of-hanoi/internal/infrastructure/persistance/postgresql"
	"github.com/AnruKitakaze/tower-of-hanoi/internal/interface/cli"
)

// TODO: Should put it into internal/.../game_handlers.go or something
type TerminalColor string

const (
	Reset   TerminalColor = "\033[0m"
	Black   TerminalColor = "\033[30m"
	Red     TerminalColor = "\033[31m"
	Green   TerminalColor = "\033[32m"
	Yellow  TerminalColor = "\033[33m"
	Blue    TerminalColor = "\033[34m"
	Magenta TerminalColor = "\033[35m"
	Cyan    TerminalColor = "\033[36m"
	White   TerminalColor = "\033[37m"
)

func PrintPeg(peg *domain.Peg) {
	d := peg.TopDisk
	if d == nil {
		fmt.Println(Red, "empty peg", Reset)
		return
	}
	fmt.Print(Reset)
	for d != nil {
		fmt.Printf("%d ", d.Size)
		d = d.Next
	}
	fmt.Println()
}

func PrintField(field *domain.Game) {
	t := make([][]uint, len(field.Pegs))
	for i := range len(t) {
		t[i] = make([]uint, field.TotalDisks)
	}

	var j int
	for i, p := range field.Pegs {
		d := p.TopDisk
		for d != nil {

			t[i][field.TotalDisks-j-1] = d.Size
			j++

			d = d.Next
		}
		j = 0
	}

	for j := field.TotalDisks - 1; j >= 0; j-- {
		for i := range len(field.Pegs) {
			if t[i][j] != 0 {
				fmt.Printf("%d\t", t[i][j])
			} else {
				fmt.Printf("|\t")
			}
		}
		fmt.Println()
	}

	for i := range len(field.Pegs) {
		fmt.Printf("Peg #%d\t", i)
	}
	fmt.Println()
}

func play(d *CliDependencies) {
	if d.playerRepo == nil {
		panic("player repository is not connected")
	}

	player, err := handleLogin(d)
	if err != nil {
		log.Fatal("play: unhandled error: %w", err)
	}

	field, err := domain.NewGame(3, 5, player, domain.DefaultColorPicker())
	if err != nil {
		panic(err)
	}

	var input []string

	fmt.Fprintln(d.out, cli.Welcome)
	PrintField(field)

	for {
		fmt.Println(Reset)
		d.scanner.Scan()
		input = strings.Split(d.scanner.Text(), " ")
		if len(input) == 1 && input[0] == "" || len(input) == 0 {
			fmt.Println(Red)
			fmt.Fprintln(d.out, cli.EmptyInput)
			continue
		} else if strings.ToLower(input[0]) == "m" {
			fmt.Println(Red)
			handleMove(d.out, input, field)
		} else if strings.ToLower(input[0]) == "h" {
			fmt.Println(Green)
			fmt.Fprintln(d.out, cli.Manual)
			continue
		} else if strings.ToLower(input[0]) == "q" {
			fmt.Println(Green)
			fmt.Fprintln(d.out, cli.Bye)
			return
		} else if strings.ToLower(input[0]) == "p" {
			fmt.Println(Yellow)
			handleGetAllPlayers(d)
			continue
		} else if strings.ToLower(input[0]) == "l" {
			fmt.Println(Yellow)
			p, err := handleLogin(d)
			if err != nil {
				fmt.Println(fmt.Errorf("failed to login: %w", err))
			} else {
				player = p

				// TODO: to func
				field, err = domain.NewGame(3, 5, player, domain.DefaultColorPicker())
				if err != nil {
					panic(err)
				}
			}
		} else if strings.ToLower(input[0]) == "r" {
			fmt.Println(Blue)
			fmt.Fprintln(d.out, "Records table is in development")
			// handleRecords(d.out, input, field)
			continue
		} else if strings.ToLower(input[0]) == "n" {
			// TODO: to func
			field, err = domain.NewGame(3, 5, player, domain.DefaultColorPicker())
			if err != nil {
				panic(err)
			}
		}

		fmt.Print(Reset)
		PrintField(field)
	}
}

func PrintPlayerInfo(w io.Writer, p *domain.Player) {
	fmt.Fprintf(w, "ID:\t%d\nNick:\t%s\n", p.ID, p.Nickname)
}

func handleGetAllPlayers(d *CliDependencies) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	players, err := d.playerRepo.GetAll(ctx)
	if err != nil {
		fmt.Fprintf(d.out, "cannot get players: %s", err.Error())
	}

	for _, p := range players {
		PrintPlayerInfo(d.out, p)
	}
}

func handleLogin(d *CliDependencies) (*domain.Player, error) {
	var player *domain.Player
	var err error

	for {
		fmt.Print(Reset)
		fmt.Print("Enter your name: ")
		d.scanner.Scan()
		nickname := d.scanner.Text()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		player, err = d.playerRepo.GetByNickname(ctx, nickname)
		if err != nil {
			if !errors.Is(err, domain.ErrPlayerNotFound) {
				log.Fatal("unable to get player by nickname: %w", err)
			}

			fmt.Fprintf(d.out, "Player with name %s does not exist. Want to create? (y/n) ", nickname)
			d.scanner.Scan()
			i := d.scanner.Text()

			saidNo := len(i) == 0 || (i != "y" && i != "д")
			if saidNo {
				continue
			}

			id, err := d.playerRepo.Save(context.Background(), nickname)
			if err != nil {
				log.Fatal("registration failed: ", err)
			}
			player, err = d.playerRepo.GetByID(context.Background(), id)
			if err != nil {
				log.Fatal("cannot find newely registered player: %w", err)
			}
		}

		return player, nil
	}
}

func handleMove(out io.Writer, input []string, field *domain.Game) {
	if len(input) < 3 {
		fmt.Fprint(out, "Swap command require two peg numbers\n")
		return
	}

	x, err := strconv.Atoi(input[1])
	if err != nil {
		fmt.Fprintf(out, "Seems like X is not a number: %v\n", err)
		return
	}

	y, err := strconv.Atoi(input[2])
	if err != nil {
		fmt.Fprintf(out, "Seems like Y is not a number: %v\n", err)
		return
	}

	if x == y {
		fmt.Fprintf(out, "X cannot be equal to Y\n")
		return
	}
	if x > len(field.Pegs) || y > len(field.Pegs) || x < 0 || y < 0 {
		fmt.Fprintf(out, "X and Y should be in a range [0, %d)\n", len(field.Pegs))
		return
	}
	err = field.MoveDisk(x, y)
	if err != nil {
		fmt.Fprintf(out, "cannot move disk: %v\n", err.Error())
	}

	if field.IsWon() {
		fmt.Fprintf(out, "Congratulations, %s! You've won! Steps: %d\n", field.Player.Nickname, field.Step)
	}
}

type CliDependencies struct {
	logger     *slog.Logger
	out        io.Writer
	scanner    *bufio.Scanner
	playerRepo domain.PlayerRepository
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	out := os.Stdout

	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	var err error

	// var playersRepo domain.PlayerRepository = inmemory.NewPlayerInmemoryRepo(logger)
	db, err := sql.Open(postgresql.DriverName, postgresql.DSN)
	if err != nil {
		logger.Error("cannot open db driver", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()

	err = postgresql.RunMigrations(logger, db)
	if err != nil {
		logger.Error("failed to apply migrations", slog.Any("err", err))
		os.Exit(1)
	}

	var playersRepo domain.PlayerRepository
	playersRepo, err = postgresql.NewPlayerPostgresRepo(logger, db)
	if err != nil {
		log.Fatal("cannot create player repo: ", err)
	}

	deps := CliDependencies{
		logger:     logger,
		out:        out,
		scanner:    scanner,
		playerRepo: playersRepo,
	}

	play(&deps)
}
