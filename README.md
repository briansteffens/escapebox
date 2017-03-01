escapebox
=========

This is a hack around [termbox-go](https://github.com/nsf/termbox-go) which
allows detection of custom escape sequences.

# Example

A working example is available here:
[examples/shifttab.go](examples/shifttab.go)

# Usage

I tried to keep the API pretty similar to using termbox directly. The basic
idea is:

1. Initialize escapebox.
2. Register any custom escape sequences you want detected.
3. Poll events with escapebox.PollEvent() instead of termbox.PollEvent().
4. Check the escapebox.Event.Seq field to see if a sequence was detected.
5. Close escapebox to clean it up.

Taking a closer look, here's how to download escapebox:

```bash
go get https://github.com/briansteffens/escapebox
```

In a termbox application, here's how to initialize escapebox:

```go
escapebox.Init()
defer escapebox.Close()
```

First you'll probably want some way to tell your sequences apart:

```go
const (
	SeqShiftTab = 1
)
```

Now you can register your custom sequences. The following means that an escape
key followed in rapid succession (< 1 ms) by two events with
termbox.Event.Chr = 91 and 90 will map to a SeqShiftTab:

```go
escapebox.Register(SeqShiftTab, 91, 90)
```

Now you can poll for events:

```go
ev := escapebox.PollEvent()
```

The escapebox.Event structure returned by escapebox.PollEvent() is identical to
the standard termbox.Event structure, except for one added field: an int called
Seq. If Seq is set to SeqNone, then it's a standard event and all the usual
fields like Ch and Key are set like normal. If Seq is not set to SeqNone, then
it's a custom escape sequence matching one of the sequences you registered:

```
if ev.Seq == SeqShiftTab {
	// The shift+tab sequence was detected
} else if ev.Seq == SeqNone {
	// This is a regular event. Check Type, Ch, and Key to see what it is
	// as if it was a standard termbox event.
}
```
