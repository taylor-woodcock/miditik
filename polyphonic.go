package main

import (
	"fmt"
)

// PolyphonicBeeper implements Beeper using multiple sources
type PolyphonicBeeper struct {
	beepers map[Beeper]bool
	// beeperMap maps a Beeper to a key for reuse
	beeperMap map[int]Beeper
}

// NewPolyphonicBeeper creates a multi-voice Beeper
func NewPolyphonicBeeper(beepers map[Beeper]bool) (Beeper, error) {
	return &PolyphonicBeeper{
		beepers:   beepers,
		beeperMap: make(map[int]Beeper),
	}, nil
}

// Beep calls Beep on an unused beeper voice and adds it to the beeperMap map
func (b *PolyphonicBeeper) Beep(keyIndex int, bend int) error {
	fmt.Printf("PolyphonicBeeper: Beep: (key: %d, bend: %d)\n", keyIndex, bend)

	var beeper Beeper

	// check if key has been assigned a Beeper
	if v, ok := b.beeperMap[keyIndex]; !ok {
		// iterate beepers and find an unused one
		for ub, used := range b.beepers {
			if !used {
				b.beeperMap[keyIndex] = beeper
				beeper = ub
			}
		}
	} else {
		// set Beeper to the one pre-assigned to this key
		beeper = v
	}

	return beeper.Beep(keyIndex, bend)
}

// NoBeep calls NoBeep on all beepers
func (b *PolyphonicBeeper) NoBeep(keyIndex int) error {
	fmt.Println("PolyphonicBeeper: NoBeep")

	// get beeper to clear
	if v, ok := b.beeperMap[keyIndex]; ok {
		err := v.NoBeep(keyIndex)
		if err != nil {
			return fmt.Errorf("hit a bad Beeper: %v", err)
		}

		// remove from beeperKeyMap and set as unused
		delete(b.beeperMap, keyIndex)
		b.beepers[v] = false
	}

	return nil
}
