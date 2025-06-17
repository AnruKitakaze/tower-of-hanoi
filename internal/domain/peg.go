package domain

import "fmt"

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
