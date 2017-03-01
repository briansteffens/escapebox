package main

import (
	"time"
	"errors"
	"github.com/nsf/termbox-go"
)


// Non-standard escape sequences
const (
	SeqNone     = 0
)


// Custom version of termbox.Event with extra data for non-standard sequences.
type Event struct {
	termbox.Event
	Seq int
}


type SequenceMask struct {
	Chars []rune
	Seq   int
}


var sequenceMasks []SequenceMask


func Register(seq int, chars ...rune) {
	sequenceMasks = append(sequenceMasks, SequenceMask {
		Seq:   seq,
		Chars: chars,
	})
}


// Convert a termbox.Event to an escapebox.Event
func makeEvent(e termbox.Event) Event {
	var ret Event

	ret.Type   = e.Type
	ret.Mod    = e.Mod
	ret.Key    = e.Key
	ret.Ch     = e.Ch
	ret.Width  = e.Width
	ret.Height = e.Height
	ret.Err    = e.Err
	ret.MouseX = e.MouseX
	ret.MouseY = e.MouseY
	ret.N      = e.N
	ret.Seq    = SeqNone

	return ret
}


// Check a list of termbox Events to see if they match any known non-standard
// escape sequences
func detectSequence(events []Event) (Event, error) {
	maskLoop: for _, mask := range sequenceMasks {
		if len(mask.Chars) + 1 != len(events) {
			continue
		}

		for i := 0; i < len(mask.Chars); i++ {
			if mask.Chars[i] != events[i + 1].Ch {
				continue maskLoop
			}
		}

		return Event { Seq: mask.Seq }, nil
	}

	return Event{}, errors.New("No sequence found.")
}


// Output channel of escapebox.Event instances
var Events chan Event


// Initialize escapebox.
func Init() {
	sequenceMasks = make([]SequenceMask, 0)

	channelizerOut = make(chan termbox.Event)
	channelizerPolling = false
	channelizerTerminate = false
	go channelizer()

	Events = make(chan Event)
	go sequencer()
}


// Read a single escapebox.Event from the output channel.
func PollEvent() Event {
	return <-Events
}


// Cleanup escapebox.
func Close() {
	// Terminate the channelizer.
	channelizerTerminate = true

	// If the termbox.PollEvent() is being called, interrupt it.
	if channelizerPolling {
		termbox.Interrupt()
	// Otherwise, read from the channel to wake it up.
	} else {
		select {
		case <-channelizerOut:
		default:
		}
	}

	// Consume any remaining events in the output to end the sequencer.
	for _ = range Events {}
}


var channelizerOut chan termbox.Event
var channelizerPolling bool
var channelizerTerminate bool


// The channelizer goroutine converts repeated termbox.PollEvent() calls to a
// channel of escapebox.Event objects.
func channelizer() {
	defer close(channelizerOut)

	for {
		if channelizerTerminate {
			break
		}

		channelizerPolling = true
		ev := termbox.PollEvent()
		channelizerPolling = false

		if channelizerTerminate {
			break
		}

		channelizerOut <- ev
	}
}


// Read termbox events and try to detect non-standard escape sequences
func sequencer() {
	defer close(Events)

	escapeSequenceMaxDuration := time.Millisecond

	// TODO: some kind of nil timer to start?
	sequenceTimer := time.NewTimer(escapeSequenceMaxDuration)
	<-sequenceTimer.C

	inEscapeSequence := false

	var buffer [10]Event
	bufferLen := 0

	for {
		select {
		case e, ok := <-channelizerOut:
			if !ok {
				return
			}

			ev := makeEvent(e)

			if ev.Type == termbox.EventKey &&
			   ev.Key == termbox.KeyEsc {
				// If already in escape sequence and we see
				// another escape key, flush the existing
				// buffer and start a new escape sequence.
				if inEscapeSequence {
					// Flush buffer
					for i := 0; i < bufferLen; i++ {
						Events <- buffer[i]
					}
				}

				sequenceTimer.Reset(escapeSequenceMaxDuration)
				inEscapeSequence = true
				bufferLen = 0
			}

			if inEscapeSequence {
				buffer[bufferLen] = ev
				bufferLen++

				seq, err := detectSequence(buffer[0:bufferLen])

				if err == nil {
					// If an escape sequence was detected,
					// return it and stop the timer.
					Events <- seq
					bufferLen = 0
					sequenceTimer.Stop()
					inEscapeSequence = false
				}

				break
			}

			// Not in possible escape sequence: handle event
			// immediately.
			Events <- ev

		case <-sequenceTimer.C:
			// Escape sequence timeout reached. Assume no escape
			// sequence is coming. Flush buffer.
			inEscapeSequence = false

			// Flush buffer
			for i := 0; i < bufferLen; i++ {
				Events <- buffer[i]
			}
		}
	}
}
