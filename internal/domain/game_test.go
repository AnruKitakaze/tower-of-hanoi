package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	tests := []struct {
		name    string
		pegs    uint
		disks   uint
		Player  *Player
		want    *Game
		wantErr error
	}{
		{
			name:    "get empty field",
			pegs:    1,
			disks:   1,
			Player:  &Player{},
			wantErr: nil,
		},
		{
			name:    "err if no player",
			pegs:    1,
			disks:   1,
			Player:  nil,
			wantErr: ErrPlayerCannotBeNil,
		},
		{
			name:    "err if no disks",
			pegs:    1,
			disks:   0,
			Player:  &Player{},
			wantErr: ErrNoDisks,
		},
		{
			name:    "err if no pegs",
			pegs:    0,
			disks:   1,
			Player:  &Player{},
			wantErr: ErrNoPegs,
		},
		{
			name:    "player ok",
			pegs:    1,
			disks:   1,
			Player:  &Player{ID: 1, Nickname: "Bobbb"},
			wantErr: nil,
		},
		{
			name:    "1 peg 1 disk",
			pegs:    1,
			disks:   1,
			Player:  &Player{},
			wantErr: nil,
		},
		{
			name:    "100 peg 1000 disk",
			pegs:    100,
			disks:   1000,
			Player:  &Player{ID: 124, Nickname: "Danno"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGame(tt.pegs, tt.disks, tt.Player, DefaultColorPicker())
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected err: %v, got %v", tt.wantErr, err)
			}
			if err != nil {
				// skip validation if any error occured
				return
			}

			validateGame(t, tt.pegs, tt.disks, tt.Player, got)
		})
	}
}

func validateGame(t *testing.T, wantPegs uint, wantDisks uint, player *Player, field *Game) {
	t.Helper()

	assert.NotNil(t, field, "field must not be nil")
	assert.Equal(t, field.Player, player, fmt.Sprintf("want player %+v, got player %+v", *player, *field.Player))
	assert.Zero(t, field.Step, "step must be zero upon creation")
	assert.NotNil(t, field.Pegs, "pegs cannot be nil")
	assert.NotEqual(t, len(field.Pegs), wantPegs, fmt.Sprintf("want %d pegs, got %d", wantPegs, len(field.Pegs)))

	sizeSet := make(map[int]struct{}, wantDisks)
	totalDisks := 0
	for p := range wantPegs {
		pegDisksCount := 0
		d := field.Pegs[p].TopDisk
		for d != nil {
			if _, ok := sizeSet[int(d.Size)]; ok == true {
				t.Errorf("duplicate Disk.Size")
			}
			if d.Next != nil {
				if d.Size > d.Next.Size {
					t.Errorf("top disk cannot be bigger than bottom one")
				}
			}
			sizeSet[int(d.Size)] = struct{}{}

			totalDisks++
			pegDisksCount++
			d = d.Next
		}

		if pegDisksCount != int(field.Pegs[p].totalDisks) {
			t.Errorf("pegs disks count error, expected %d, got %d", field.Pegs[p].totalDisks, pegDisksCount)
		}
	}

	if totalDisks != int(wantDisks) {
		t.Errorf("want %d disks, got %d", wantDisks, totalDisks)
	}
}
