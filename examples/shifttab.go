package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)


// Custom sequence identifiers - will be used in escapebox.Event.Seq to tell
// apart sequences from normal events. If escapebox.Event.Seq is SeqNone,
// the event is not a custom sequence.
const (
	SeqShiftTab = 1
)


// Convenience function to write a formatted string to termbox output.
func termPrintf(x, y int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	for i, c := range s {
		termbox.SetCell(x + i, y, c, termbox.ColorWhite,
				termbox.ColorBlack)
	}
}


func main() {
	// Initialize termbox normally.
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	// Initialize escapebox.
	escapebox.Init()
	defer escapebox.Close()

	// Register any custom sequences. An escape key followed by two events
	// with Ch = 91 and Ch = 90 will be interpreted as a SeqShiftTab.
	escapebox.Register(SeqShiftTab, 91, 90)

	for {
		// Poll an event through escapebox rather than termbox to get
		// sequence detection.
		ev := escapebox.PollEvent()

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		if ev.Seq == SeqShiftTab {
			termPrintf(5, 5, "Shift+Tab")
		} else if ev.Key == termbox.KeyCtrlC {
			break
		} else {
			termPrintf(5, 5, "%d %d", ev.Key, ev.Ch)
		}

		termbox.Flush()
		termbox.Sync()
	}
}
