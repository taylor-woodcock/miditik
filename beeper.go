package main

import (
	"fmt"
	"math"

	ssh "github.com/helloyi/go-sshclient"
)

const (
	InitBeepFrequency  = 100
	InitBeepDuration   = 1
	ClearBeepFrequency = 20
	ClearBeepDuration  = 0.001
	BendMultiplier     = 1.0
)

// Beeper defines a beepable device
type Beeper interface {
	Beep(keyIndex int, bend int) error
	NoBeep() error
}

// MikroTikBeeper implements Beeper using MikroTik router
type MikroTikBeeper struct {
	client  *ssh.Client
	midiMap map[int]float64
}

// NewMikroTikBeeper creates a MikroTik-backed Beeper
func NewMikroTikBeeper(client *ssh.Client, midiMap map[int]float64) (Beeper, error) {
	err := beep(client, InitBeepFrequency, InitBeepDuration)
	if err != nil {
		return nil, err
	}

	return &MikroTikBeeper{
		client:  client,
		midiMap: midiMap,
	}, nil
}

// Beep calculates the frequency and runs a beep command on the host
func (b *MikroTikBeeper) Beep(keyIndex int, bend int) error {
	fmt.Printf("Beep - Key: %d, bend: %d\n", keyIndex, bend)
	frequency := math.Pow(2, ((float64(keyIndex)-69)/12.0)+((float64(bend)-8192)/(4096*12))) * 440
	return beep(b.client, frequency, BeepDuration)
}

// NoBeep runs a clearing beep command to stop the beeping
func (b *MikroTikBeeper) NoBeep() error {
	fmt.Println("NoBeep")
	return beep(b.client, ClearBeepFrequency, ClearBeepDuration)
}

// beep runs a beep command with a defined frequency and duration on the client
func beep(client *ssh.Client, frequency float64, duration float64) error {
	cmd := fmt.Sprintf("beep frequency=%f length=%f", frequency, duration)
	fmt.Printf("Cmd: %s\n", cmd)
	return client.Cmd(cmd).Run()
}
