package main

import (
	"fmt"
	"math"

	ssh "github.com/helloyi/go-sshclient"
)

const (
	// NoBeepFrequency determines the frequency sent when NoBeep is called
	NoBeepFrequency = 20
	// NoBeepDuration determines the duration of the beep when NoBeep is called
	NoBeepDuration = 0.001
)

// MikroTikBeeper implements Beeper using MikroTik router source
type MikroTikBeeper struct {
	client  *ssh.Client
	midiMap map[int]float64
}

// NewMikroTikBeeper creates a MikroTik-backed Beeper
func NewMikroTikBeeper(client *ssh.Client, midiMap map[int]float64) (Beeper, error) {
	b := &MikroTikBeeper{
		client:  client,
		midiMap: midiMap,
	}

	return b, nil
}

// Beep calculates the frequency and runs a beep command on the host
//
// https://dsp.stackexchange.com/questions/1645/converting-a-pitch-bend-midi-value-to-a-normal-pitch-value
func (b *MikroTikBeeper) Beep(keyIndex int, bend int) error {
	fmt.Printf("MikroTikBeeper: Beep: (key: %d, bend: %d) - Frequency: ", keyIndex, bend)

	frequency := math.Pow(2, ((float64(keyIndex)-69)/12.0)+((float64(bend)-8192)/(4096*12))) * 440

	fmt.Println(frequency)

	return beep(b.client, frequency, BeepDuration)
}

// NoBeep runs a clearing beep command to stop the beeping
func (b *MikroTikBeeper) NoBeep(keyIndex int) error {
	fmt.Println("MikroTikBeeper: NoBeep")

	return beep(b.client, NoBeepFrequency, NoBeepDuration)
}

// beep runs a beep command with a defined frequency and duration on the client
func beep(client *ssh.Client, frequency float64, duration float64) error {
	cmd := fmt.Sprintf("beep frequency=%f length=%f", frequency, duration)

	return client.Cmd(cmd).Run()
}
