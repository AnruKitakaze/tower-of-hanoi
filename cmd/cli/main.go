package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/AnruKitakaze/tower-of-hanoi/internal/domain"
	"github.com/AnruKitakaze/tower-of-hanoi/internal/text"
)

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

func PrintField(field *domain.Field) {
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

func play(s *bufio.Scanner, out io.Writer) {
	field, err := domain.NewGame(3, 5, &domain.Player{ID: 1, Nickname: "Anru"}, domain.DefaultColorPicker())
	if err != nil {
		panic(err)
	}

	var input []string

	fmt.Fprintln(out, text.Welcome)
	PrintField(field)

	for {
		fmt.Println(Reset)
		s.Scan()
		input = strings.Split(s.Text(), " ")
		if len(input) == 1 && input[0] == "" || len(input) == 0 {
			fmt.Println(Red)
			fmt.Fprintln(out, text.EmptyInput)
			continue
		} else if strings.ToLower(input[0]) == "m" {
			fmt.Println(Red)
			handleMove(out, input, field)
		} else if strings.ToLower(input[0]) == "h" {
			fmt.Println(Green)
			fmt.Fprintln(out, text.Manual)
			continue
		} else if strings.ToLower(input[0]) == "q" {
			fmt.Println(Green)
			fmt.Fprintln(out, text.Bye)
			return
		} else if strings.ToLower(input[0]) == "r" {
			fmt.Println(Blue)
			fmt.Fprintln(out, "Records table is in development")
			continue
		} else if strings.ToLower(input[0]) == "n" {
			field, err = domain.NewGame(3, 5, &domain.Player{ID: 1, Nickname: "Anru"}, domain.DefaultColorPicker())
			if err != nil {
				panic(err)
			}
		}

		fmt.Print(Reset)
		PrintField(field)
		if field.IsWon() {
			fmt.Fprintf(out, "Congratulations, %s! You've won! Steps: %d\n", field.Player.Nickname, field.Step)
		}
	}
}

func handleMove(out io.Writer, input []string, field *domain.Field) {
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
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	out := os.Stdout

	play(scanner, out)
}
