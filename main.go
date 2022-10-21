package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/shiftregister"
)

// Digit bits
//
//       7
//       v
//       -
//  2-> | | <- 6
//       -  <- 1
//  3-> | | <- 5
//       -    . <- 0
//       ^
//       4

//
// The bitmap for each LED digit.
// Designed by Pepper in Melbourne.
// The bit 0 is the decimal point.
//
var digits = []uint32{
	0b1111110_0, // 0
	0b0110000_0, // 1
	0b1101101_0, // 2
	0b1111001_0, // 3
	0b0110011_0, // 4
	0b1011011_0, // 5
	0b0011111_0, // 6
	0b1110000_0, // 7
	0b1111111_0, // 8
	0b1110011_0, // 9
}

const (
	ButtonOff = iota
	ButtonPressed
	ButtonHeld
)

// The state of each LED digit.
var state = []uint32{0, 0, 0, 0}

// The state of the dots. This is managed separately.
var dots = []uint32{0, 0, 0, 0}

// Button ID and state
type button struct {
	index   int
	pin     machine.Pin
	dwell   int
	pressed bool // set only when the button is first pressed
}

// Buttons used to control the clock
var buttons []*button

// The amount that's been added/subtracted from the time by the adjustment buttons
var offset int

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	device := shiftregister.New(12, machine.Pin(9), machine.Pin(10), machine.Pin(11))
	device.Configure()

	// Buttons for setting the time.
	addButton(16) // First button updates the minute
	addButton(17) // Second button updates the hour

	// sentinal flashing led
	go sentinal(led)

	// Update the time
	go clock()

	// Check for buttons
	go checkButtons()

	render(device)
}

func addButton(pinId int) *button {
	pin := machine.Pin(pinId)
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	b := &button{index: len(buttons), pin: pin, dwell: 0}
	buttons = append(buttons, b)
	return b
}

func render(device *shiftregister.Device) {
	//
	// This is the main loop. It simply writes the bitmaps out
	// to illuminate one LED at a time. Do it fast enough, and
	// you can't see it!
	//
	for {
		for led := uint32(0); led < 4; led++ {
			bits := state[led] | dots[led]
			device.WriteMask(bits<<4 | 0b1111 ^ (1 << led))
			time.Sleep(time.Millisecond * 1)
		}
	}
}

// Update the button state based on the GPIO.
func (b *button) updateState() {
	if b.pin.Get() {
		b.pressed = (b.dwell == 0) // set only the first time button down is detected
		b.dwell++
		dots[b.index] = 1 // DEBUGGING: show which button is pressed
	} else {
		b.dwell = 0
		b.pressed = false
		dots[b.index] = 0 // DEBUGGING: show which button is pressed
	}
}

func (b *button) state() int {
	if b.pressed {
		return ButtonPressed
	}

	if b.dwell > 50 {
		return ButtonHeld
	}

	return ButtonOff
}

func checkButtons() {
	for {
		// Keep track of how long a button has been held down.
		for _, button := range buttons {
			button.updateState()
		}

		updateTime()

		time.Sleep(time.Millisecond * 50)
	}
}

// Check the state of the buttons and update the time accordingly.
func updateTime() {
	hour := offset / 3600 % 24
	minute := offset / 60 % 60

	switch buttons[0].state() {
	case ButtonPressed:
		minute++
		if minute > 59 {
			minute = 0
		}
		setOffset(hour, minute)
	case ButtonHeld:
		minute++
		if minute > 59 {
			minute = 0
		}
		setOffset(hour, minute)
	}

	switch buttons[1].state() {
	case ButtonPressed:
		hour++
		if hour > 23 {
			hour = 0
		}
		setOffset(hour, minute)
	}
}

func setOffset(hour int, minute int) {
	offset = hour*3600 + minute*60 // reset seconds to zero
	updateFace()
}

//
// This is blinky. It simply displays the green flashing LED
// to show that we're alive. It also sets the seconds LED in
// the front of the display.
//
func sentinal(led machine.Pin) {

	isSet := true

	for {
		time.Sleep(time.Millisecond * 500)
		if isSet {
			led.High()
			dots[2] = 1
		} else {
			led.Low()
			dots[2] = 0
		}

		isSet = !isSet
	}
}

func clock() {
	for {
		updateFace()
		time.Sleep(time.Second)
	}
}

// Converts the time in seconds to clock-face time.
func updateFace() {

	now := time.Now().Unix()
	when := int(now) + offset

	// Convert to minutes in the day
	minute := when / 60 % 60
	hour := when / 3600 % 24

	// 12 hour time
	// if hour == 0 {
	// 	hour = 12
	// }

	Display(hour*100 + minute)
}

//
// Display the given number.
//
func Display(number int) {
	led := 0

	for number > 0 && led < 4 {
		digit := number % 10
		state[led] = digits[digit]
		led++
		number = number / 10
	}

	// Clear any digits we didn't explicitly set.
	for led < 4 {
		state[led] = digits[0]
		led++
	}
}
