package main

import (
	"fmt"
	"sort"
)

// MidiBeeper implements Beeper that handles Midi values
//
// This Beeper is capable of processing key press and bend commands independently.
//
// It will forward the Beep requests on when necessary, which is when the bend diifference
// is reaches a threshold or a key is pressed.
//
// Keys are also played in order of midi key index, meaning the highest key will be played
type MidiBeeper struct {
	beeper      Beeper
	pressedKeys []int
	bend        int
}

// NewMidiBeeper creates a MidiBeeper that can process MIDI values
func NewMidiBeeper(beeper Beeper) (Beeper, error) {
	return &MidiBeeper{
		beeper:      beeper,
		pressedKeys: make([]int, 0),
		bend:        BendZero,
	}, nil
}

// Beep decodes raw key and bend values and calls Beep on the child Beeper when necesary
func (b *MidiBeeper) Beep(keyIndex int, bend int) (err error) {
	fmt.Printf("MidiBeeper: Beep(k: %d, bend: %d)\n", keyIndex, bend)
	b.bend = bend

	var last int
	if keyIndex > 0 {
		b.pressedKeys = append(b.pressedKeys, keyIndex)
	}

	if len(b.pressedKeys) > 0 {
		sort.Ints(b.pressedKeys)
		last = b.pressedKeys[len(b.pressedKeys)-1]
	}

	if last > 0 {
		err = b.beeper.Beep(last, bend)
	}

	return err
}

// NoBeep decides whether to call NoBeep when there are no remaining beeps, or call Beep using the previous key
func (b *MidiBeeper) NoBeep(keyIndex int) (err error) {
	fmt.Print("MidiBeeper: NoBeep - Remaining: ")

	sort.Ints(b.pressedKeys)
	b.pressedKeys, _ = remove(b.pressedKeys, keyIndex)

	fmt.Println(b.pressedKeys)

	if len(b.pressedKeys) > 0 {
		err = b.beeper.Beep(b.pressedKeys[len(b.pressedKeys)-1], b.bend)
	} else {
		err = b.beeper.NoBeep(keyIndex)
	}
	return err
}
