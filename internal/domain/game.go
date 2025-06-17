package domain

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand"
)

var ErrPlayerCannotBeNil = errors.New("player cannot be nil")
var ErrNoPegs = errors.New("pegs count cannot be < 1")
var ErrNoDisks = errors.New("disks count cannot be < 1")

// Game contain all information about current gaming session
type Game struct {
	Pegs       []Peg
	TotalDisks int
	Step       uint
	Player     *Player
}

func (g *Game) MoveDisk(fromPeg int, toPeg int) error {
	if fromPeg < 0 || toPeg < 0 || fromPeg >= g.TotalDisks || toPeg >= g.TotalDisks {
		return fmt.Errorf("fromPeg and toPeg should be in range [0, %d)", g.TotalDisks)
	}

	if g.Pegs[fromPeg].TopDisk != nil && g.Pegs[toPeg].TopDisk != nil && g.Pegs[fromPeg].TopDisk.Size > g.Pegs[toPeg].TopDisk.Size {
		return fmt.Errorf("cannot put bigger disk on top of smaller one")
	}

	d, err := g.Pegs[fromPeg].GrabDisk()
	if err != nil {
		return fmt.Errorf("cannot grab disk: %w", err)
	}

	err = g.Pegs[toPeg].PutDisk(d)
	if err != nil {
		return fmt.Errorf("cannot put disk: %w", err)
	}

	g.Step++
	return nil
}

// TODO: Должно использоваться тут... usecase?
func (g *Game) IsWon() bool {
	idx := -1
	for i, p := range g.Pegs {
		if p.totalDisks > 0 {
			if idx != -1 {
				return false
			}
			idx = i
		}
	}

	for d := g.Pegs[idx].TopDisk; d.Next != nil; d = d.Next {
		if d.Size > d.Next.Size {
			return false
		}
	}

	return true
}

func DefaultColorPicker() func() color.Color {
	c := -1
	return func() color.Color {
		c++
		return diskColors[c%len(diskColors)]
	}
}

func NewGame(pegs uint, disks uint, player *Player, colorPicker func() color.Color) (*Game, error) {
	if player == nil {
		return nil, ErrPlayerCannotBeNil
	}

	if pegs < 1 {
		return nil, ErrNoPegs
	}

	if disks < 1 {
		return nil, ErrNoDisks
	}

	p := make([]Peg, pegs)

	for i := range disks {
		curr := &Disk{
			Size:  disks - i,
			Color: colorPicker(),
			Next:  nil,
		}

		pegIdx := rand.Intn(int(pegs))
		if p[pegIdx].TopDisk != nil {
			curr.Next = p[pegIdx].TopDisk
		}
		p[pegIdx].TopDisk = curr
		p[pegIdx].totalDisks++
	}

	g := &Game{
		Pegs:       p,
		TotalDisks: int(disks),
		Step:       0,
		Player:     player,
	}

	return g, nil
}
