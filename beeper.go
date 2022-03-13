package main

// Beeper defines a beepable device
type Beeper interface {
	Beep(keyIndex int, bend int) error
	NoBeep(keyIndex int) error
}
