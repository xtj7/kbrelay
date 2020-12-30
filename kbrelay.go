package main

import (
	"github.com/MarinX/keylogger"
	"github.com/sirupsen/logrus"
)

func main() {
	forwardKeys := false

	// find keyboard device, does not require a root permission
	keyboard := keylogger.FindKeyboardDevice()

	logrus.Println("Found a keyboard at", keyboard)
	// init keylogger with keyboard
	k, err := keylogger.New(keyboard)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer k.Close()

	events := k.Read()

	enabledKeys := [256]bool{}
	for i := 0; i < 255; i++ {
		enabledKeys[0] = false
	}

	// range of events
	for e := range events {
		switch e.Type {
		// EvKey is used to describe state changes of keyboards, buttons, or other key-like devices.
		// check the input_event.go for more events
		case keylogger.EvKey:

			// if the state of key is pressed
			if e.KeyPress() {
				logrus.Println("[event] press key ", e.KeyString(), e.Code)
				enabledKeys[e.Code] = true
			}

			// if the state of key is released
			if e.KeyRelease() {
				logrus.Println("[event] release key ", e.KeyString(), e.Code)
				enabledKeys[e.Code] = false
			}

			if enabledKeys[188] && enabledKeys[189] {
				forwardKeys = !forwardKeys
				logrus.Println("Changed forwardKeys to %v", forwardKeys)
			}

			break
		}
	}
}
