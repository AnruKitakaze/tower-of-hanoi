package domain

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand"
)

var ErrNoPlayer = errors.New("player cannot be nil")
var ErrNoPegs = errors.New("pegs count cannot be < 1")
var ErrNoDisks = errors.New("disks count cannot be < 1")

var (
	Red    color.Color = color.RGBA{255, 0, 0, 100}
	Orange color.Color = color.RGBA{255, 100, 0, 100}
	Yellow color.Color = color.RGBA{255, 255, 0, 100}
	Green  color.Color = color.RGBA{0, 255, 0, 100}
	Cyan   color.Color = color.RGBA{0, 255, 150, 100}
	Blue   color.Color = color.RGBA{0, 0, 255, 100}
	Purple color.Color = color.RGBA{150, 0, 255, 100}
)

var diskColors = []color.Color{Red, Orange, Yellow, Green, Cyan, Blue, Purple}

// Rings are storeg in a peg
// Next is the disk below current one in the same peg
type Disk struct {
	Size  uint
	Color color.Color
	Next  *Disk
}

// Peg is a stick which holds disks in Tower of Hanoi
type Peg struct {
	TopDisk    *Disk
	totalDisks uint
}

func (p *Peg) GrabDisk() (*Disk, error) {
	if p.totalDisks == 0 {
		return nil, fmt.Errorf("GrabDisk: peg is empty")
	}

	d := p.TopDisk
	p.TopDisk = p.TopDisk.Next
	p.totalDisks--

	return d, nil
}

func (p *Peg) PutDisk(disk *Disk) error {
	if disk == nil {
		return fmt.Errorf("PutDisk: disk cannot be nil")
	}

	disk.Next = p.TopDisk
	p.TopDisk = disk
	p.totalDisks++

	return nil
}

// Field contain all information about current gaming session
type Field struct {
	Pegs       []Peg
	TotalDisks int
	Step       uint
	Player     *Player
}

func (f *Field) MoveDisk(fromPeg int, toPeg int) error {
	if fromPeg < 0 || toPeg < 0 || fromPeg >= f.TotalDisks || toPeg >= f.TotalDisks {
		return fmt.Errorf("fromPeg and toPeg should be in range [0, %d)", f.TotalDisks)
	}

	if f.Pegs[fromPeg].TopDisk != nil && f.Pegs[toPeg].TopDisk != nil && f.Pegs[fromPeg].TopDisk.Size > f.Pegs[toPeg].TopDisk.Size {
		return fmt.Errorf("cannot put bigger disk on top of smaller one")
	}

	d, err := f.Pegs[fromPeg].GrabDisk()
	if err != nil {
		return fmt.Errorf("cannot grab disk: %w", err)
	}

	err = f.Pegs[toPeg].PutDisk(d)
	if err != nil {
		return fmt.Errorf("cannot put disk: %w", err)
	}

	f.Step++
	return nil
}

func (f *Field) IsWon() bool {
	idx := -1
	for i, p := range f.Pegs {
		if p.totalDisks > 0 {
			if idx != -1 {
				return false
			}
			idx = i
		}
	}

	for d := f.Pegs[idx].TopDisk; d.Next != nil; d = d.Next {
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

func NewGame(pegs uint, disks uint, player *Player, colorPicker func() color.Color) (*Field, error) {
	if player == nil {
		return nil, ErrNoPlayer
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

	f := &Field{
		Pegs:       p,
		TotalDisks: int(disks),
		Step:       0,
		Player:     player,
	}

	return f, nil
}
