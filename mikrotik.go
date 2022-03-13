package main

import (
	"fmt"
	"math"
	"time"

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

	b.InitSound()

	return b, nil
}

// InitSound plays a sound on Initialisation
func (b *MikroTikBeeper) InitSound() error {
	b.Beep(60, BendZero)
	time.Sleep(time.Millisecond * 100)

	b.Beep(64, BendZero)
	time.Sleep(time.Millisecond * 100)

	b.Beep(67, BendZero)
	time.Sleep(time.Millisecond * 500)

	b.NoBeep()

	return nil
}

// Beep calculates the frequency and runs a beep command on the host
func (b *MikroTikBeeper) Beep(keyIndex int, bend int) error {
	fmt.Printf("Beep: (key: %d, bend: %d)\n", keyIndex, bend)
	frequency := math.Pow(2, ((float64(keyIndex)-69)/12.0)+((float64(bend)-8192)/(4096*12))) * 440
	return beep(b.client, frequency, BeepDuration)
}

// NoBeep runs a clearing beep command to stop the beeping
func (b *MikroTikBeeper) NoBeep() error {
	fmt.Println("NoBeep")
	return beep(b.client, NoBeepFrequency, NoBeepDuration)
}

// beep runs a beep command with a defined frequency and duration on the client
func beep(client *ssh.Client, frequency float64, duration float64) error {
	cmd := fmt.Sprintf("beep frequency=%f length=%f", frequency, duration)
	fmt.Printf("Cmd: %s\n", cmd)
	return client.Cmd(cmd).Run()
}
