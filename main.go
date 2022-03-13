package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"

	ssh "github.com/helloyi/go-sshclient"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	testdrv "gitlab.com/gomidi/midi/testdrv"
	midicat "gitlab.com/gomidi/midicatdrv"
)

const (
	// TODO Replace with cobra CLI flags
	// BeepDuration determines the max duration each beep plays for in seconds
	BeepDuration = 10
	// BendDivision determines the total diference in bend before we send another command
	BendDivision = 700
	// Driver determines the MIDI driver we're using
	Driver = Midicat
	// Host determines the SSH connection address
	Host = "192.168.88.1:22"
	// User determines the SSH username
	User = "admin"
	// Pass determines the SSH password
	Pass = ""
)

// Driver defines a midi driver name
type MidiDriver string

const (
	Test    MidiDriver = "test"
	Midicat MidiDriver = "midicat"
)

// MessageKey defines the keys of each midi message
type MessageKey string

const (
	Key       MessageKey = "key"
	Channel   MessageKey = "channel"
	Velocity  MessageKey = "velocity"
	Frequency MessageKey = "frequency"
)

// Action defines a Midi action
type Action int

const (
	Invalid Action = iota
	NoteOff
	NoteOn
	Pitchbend
)

// Midi defines a midi message
type Midi struct {
	channel   int
	key       int
	action    Action
	velocity  int
	frequency float64
	value     int
}

func main() {
	fmt.Printf("Connecting to ssh: %s\n", Host)

	client, err := ssh.DialWithPasswd(Host, User, Pass)
	if err != nil {
		must(err)
	}
	defer func() {
		fmt.Println("Closing ssh")
		client.Close()
	}()

	fmt.Println("Connection successful!")

	var pressedKeys []int
	bend := BendDefault
	var bendDif float64
	midiMap := calculateMidiFrequencies(0, 255)
	beeper, err := NewMikroTikBeeper(client, midiMap)
	// randomBeeps(beeper, midiMap)

	fmt.Printf("Connecting to midi using: %s\n", Driver)

	// select driver
	var drv midi.Driver
	switch Driver {
	case Test:
		drv = testdrv.New("MidiTik")
	case Midicat:
		drv, err = midicat.New()
	default:
		err = fmt.Errorf("invalid midi driver")
	}
	must(err)

	// ensure driver is closed
	defer func() {
		fmt.Println("Closing midi driver")
		drv.Close()
	}()

	ins, err := drv.Ins()
	must(err)
	outs, err := drv.Outs()
	must(err)

	in, out := ins[0], outs[0]
	must(in.Open())
	must(out.Open())

	defer func() {
		fmt.Println("Closing input")
		in.Close()
	}()
	defer func() {
		fmt.Println("Closing output")
		out.Close()
	}()

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			fmt.Printf("Processing midi: %s\n", msg.String())

			midi, err := decodeMidi(msg)
			if err != nil {
				fmt.Printf("failed to deocde midi: %v\n", err)
				return
			}

			fmt.Printf("Got midi: %#v\n", midi)

			switch midi.action {
			case NoteOn:
				pressedKeys = append(pressedKeys, midi.key)
				sort.Ints(pressedKeys)
				last := pressedKeys[len(pressedKeys)-1]

				if midi.key == last {
					// beep highest key
					err = beeper.Beep(last, bend)
				}
			case NoteOff:
				sort.Ints(pressedKeys)
				pressedKeys, _ = remove(pressedKeys, midi.key)

				if len(pressedKeys) > 0 {
					err = beeper.Beep(pressedKeys[len(pressedKeys)-1], bend)
				} else {
					err = beeper.NoBeep()
				}
			case Pitchbend:
				bendDif += math.Abs(float64(bend - midi.value))
				bend = midi.value

				fmt.Printf("BendDif: %f\n", bendDif)

				if len(pressedKeys) > 0 && bendDif >= BendDivision {
					err = beeper.Beep(pressedKeys[len(pressedKeys)-1], bend)
				}

				if bendDif >= BendDivision {
					bendDif = 0
				}
			default:
				err = fmt.Errorf("invalid action: %v", midi.action)
			}
			if err != nil {
				fmt.Printf("Could not process midi: %v\n", err)
			}
		}),
	)

	// listen for midi
	err = rd.ListenTo(in)
	must(err)

	fmt.Println("Connection successful!")

	// system interrupts
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// await interrupt and exit
	<-signalChan
}
