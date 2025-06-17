package domain

import "image/color"

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
