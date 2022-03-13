package main

import (
	"fmt"
	"math"
)

// LimitedBeeper implements Beeper that limits the amount of Beeps forwarded based on the bend or key.
//
// It will forward the Beep requests when the bend difference reaches a threshold or a different key is pressed
type LimitedBeeper struct {
	beeper  Beeper
	lastKey int
	bend    int
	bendDif float64
}

// NewLimitedBeeper creates a LimitedBeeper that can limits forwarded Beep calls based on bend and key
func NewLimitedBeeper(beeper Beeper, bendInterval float64) (Beeper, error) {
	return &LimitedBeeper{
		beeper:  beeper,
		bend:    BendZero,
		bendDif: bendInterval,
	}, nil
}

// Beep forwards Beep call when key changes or bend value difference threshold is reached
func (b *LimitedBeeper) Beep(keyIndex int, bend int) (err error) {
	fmt.Printf("LimitedBeeper: Beep(k: %d, bend: %d) - BendDif: %.2f\n", keyIndex, bend, b.bendDif)

	// calculate bend dif to limit calls
	b.bendDif += math.Abs(float64(b.bend - bend))
	b.bend = bend

	// fmt.Printf("BendDif: %f\n", b.bendDif)

	if keyIndex == b.lastKey && b.bendDif < BendDivision {
		return
	}

	b.lastKey = keyIndex

	err = b.beeper.Beep(keyIndex, bend)
	if err != nil {
		return err
	}

	if b.bendDif >= BendDivision {
		b.bendDif = 0
	}

	return err
}

// NoBeep decides forwards NoBeep call
func (b *LimitedBeeper) NoBeep(keyIndex int) (err error) {
	fmt.Println("LimitedBeeper: NoBeep")

	b.lastKey = 0

	return b.beeper.NoBeep(keyIndex)
}
