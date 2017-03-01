package main

import (
	"fmt"
	"os"
	"github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	f, err := os.Create("out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	Init()
	defer Close()

	for {
		ev := PollEvent()

		if ev.Seq == SeqShiftTab {
			f.WriteString("shift tab\n")
		} else if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlC {
			f.WriteString("ctrl c\n")
			break
		} else {
			f.WriteString(fmt.Sprintf("%d %d\n", ev.Key, ev.Ch))
		}
	}
}
