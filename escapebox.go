package main

import (
    "time"
    "github.com/nsf/termbox-go"
)


// Non-standard escape sequences
const (
    SeqNone     = 0
    SeqShiftTab = 1
)


// Custom version of termbox.Event with extra data for non-standard sequences.
type Event struct {
    termbox.Event
    Seq int
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
func detectSequence(events []Event) Event {
    var ret Event

    if len(events) == 3 &&
       events[0].Type == termbox.EventKey &&
       events[0].Key == termbox.KeyEsc &&
       events[1].Type == termbox.EventKey &&
       events[1].Key == 0 &&
       events[1].Ch == 91 &&
       events[2].Type == termbox.EventKey &&
       events[2].Key == 0 &&
       events[2].Ch == 90 {
        ret.Seq = SeqShiftTab
    }

    return ret
}


// Output channel of escapebox.Event instances
var Events chan Event


// Initialize escapebox.
func Init() {
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

    // Consume any remaining events in the output so the sequencer will end.
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
    escapeSequenceTimer := time.NewTimer(escapeSequenceMaxDuration)
    <-escapeSequenceTimer.C

    inEscapeSequence := false

    var sequenceBuffer [10]Event
    sequenceBufferLen := 0

    for {
        select {
        case e, ok := <-channelizerOut:
            if !ok {
                return
            }

            ev := makeEvent(e)

            if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
                // If already in escape sequence and we see another escape key,
                // flush the existing buffer and start a new escape sequence.
                if inEscapeSequence {
                    // Flush buffer
                    for i := 0; i < sequenceBufferLen; i++ {
                        Events <- sequenceBuffer[i]
                    }
                }

                escapeSequenceTimer.Reset(escapeSequenceMaxDuration)
                inEscapeSequence = true
                sequenceBufferLen = 0
            }

            if inEscapeSequence {
                sequenceBuffer[sequenceBufferLen] = ev
                sequenceBufferLen++

                seq := detectSequence(sequenceBuffer[0:sequenceBufferLen])

                if seq.Seq != SeqNone {
                    // If an escape sequence was detected, return it and stop
                    // the timer.
                    Events <- seq
                    sequenceBufferLen = 0
                    escapeSequenceTimer.Stop()
                    inEscapeSequence = false
                }

                break
            }

            // Not in possible escape sequence: handle event immediately.
            Events <- ev

        case <-escapeSequenceTimer.C:
            // Escape sequence timeout reached. Assume no escape sequence is
            // coming. Flush buffer.
            inEscapeSequence = false

            // Flush buffer
            for i := 0; i < sequenceBufferLen; i++ {
                Events <- sequenceBuffer[i]
            }
        }
    }
}
