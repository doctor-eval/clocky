# Clocky

This is a 4-digit digital clock using a raspberry pi pico and the Waveshare Pico-8SEG-LED:

    https://www.waveshare.com/wiki/Pico-8SEG-LED

Two GPIO pins are used to manually set the time; I connected them to pushbuttons on a breadboard via
a 10K resistor:

* GP16 is the minute setter
* GP17 is the hour setter

The minute setter will speed up if you hold the button down.

It includes a self-built LED driver created by reverse engineering the waveshare.
Maybe there is an existing driver for the display but I didn't know about it.

# Building

    tinygo build -target=pico

# Flashing

Hold down the button on the Pico when plugging it in, then run:

    tinygo flash -target=pico
