package main

import (
	"bytes"
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/sirupsen/logrus"
	"os"
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

	enabledKeys := make(map[string]bool)

	// open output
	f, err := os.OpenFile("/dev/hidg0", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatal(err)
	}

	// range of events
	for e := range events {
		switch e.Type {
		// EvKey is used to describe state changes of keyboards, buttons, or other key-like devices.
		// check the input_event.go for more events
		case keylogger.EvKey:
			var keyString = e.KeyString()
			if keyString == "" {
				keyString = fmt.Sprintf("KEY_%d", e.Code)
			}

			// if the state of key is pressed
			if e.KeyPress() {
				logrus.Println("[event] press key ", e.KeyString(), e.Code, e.Value)
				enabledKeys[keyString] = true
			}

			// if the state of key is released
			if e.KeyRelease() {
				logrus.Println("[event] release key ", e.KeyString(), e.Code, e.Value)
				enabledKeys[keyString] = false
			}

			if enabledKeys["KEY_188"] && enabledKeys["KEY_189"] {
				forwardKeys = !forwardKeys
				logrus.Println("Changed forwardKeys to", forwardKeys)
			}

			if forwardKeys {
				sendKeys(enabledKeys, f)
			}

			break
		}
	}
}

func sendKeys(enabledKeys map[string]bool, f *os.File) {
	keyList := []string{}
	for keyName, enabled := range enabledKeys {
		if enabled && !isModifierKey(keyName) {
			keyList = append(keyList, keyName)
		}
	}

	var buf bytes.Buffer
	if len(keyList) > 0 {
		buf.WriteRune(rune(getModifierCode(enabledKeys)))
		buf.WriteRune(rune(0))
		for i := 0; i < 6; i++ {
			if i < len(keyList) {
				buf.WriteRune(rune(keyCodeToScanCode(keyList[i])))
			} else {
				buf.WriteRune(rune(0))
			}
		}
	} else {
		// Release all keys
		for i := 0; i < 8; i++ {
			buf.WriteRune(rune(0))
		}

	}
	buf.WriteTo(f)
}

func isModifierKey(keyName string) bool {
	// check if is modifier code
	_, ok := getModifierCodeForKey(keyName)
	return ok
}

func getModifierCode(enabledKeys map[string]bool) int {
	// iterate over all keys, check for modifier keys, return bit sum of all modifier keys
	var returnModifierCode int = 0
	for keyName, enabled := range enabledKeys {
		if enabled && isModifierKey(keyName) {
			modifierCode, _ := getModifierCodeForKey(keyName)
			returnModifierCode += modifierCode
		}
	}
	return returnModifierCode
}

func getModifierCodeForKey(keyName string) (int, bool) {
	m := make(map[string]int)
	m["KEY_126"] = 0
	m["R_ALT"] = 1
	m["R_SHIFT"] = 2
	m["R_CTRL"] = 4
	m["KEY_125"] = 8
	m["L_ALT"] = 16
	m["L_SHIFT"] = 32
	m["L_CTRL"] = 64

	val, ok := m[keyName]
	return val, ok
}

func keyCodeToScanCode(keyCode string) int {
	m := make(map[string]int)
	m["A"] = 0x04
	m["B"] = 0x05
	m["C"] = 6
	m["D"] = 7
	m["E"] = 8
	m["F"] = 9
	m["G"] = 10
	m["H"] = 11
	m["I"] = 12
	m["J"] = 13
	m["K"] = 14
	m["L"] = 15
	m["M"] = 16
	m["N"] = 17
	m["O"] = 18
	m["P"] = 19
	m["Q"] = 20
	m["R"] = 21
	m["S"] = 22
	m["T"] = 23
	m["U"] = 24
	m["V"] = 25
	m["W"] = 26
	m["X"] = 27
	m["Y"] = 28
	m["Z"] = 29
	m["1"] = 30
	m["2"] = 31
	m["3"] = 32
	m["4"] = 33
	m["5"] = 34
	m["6"] = 35
	m["7"] = 36
	m["8"] = 37
	m["9"] = 38
	m["0"] = 39
	m["BS"] = 0x2a
	m["ENTER"] = 0x28
	m["ESC"] = 0x29
	m["TAB"] = 0x2b
	m["SPACE"] = 0x2c
	m[";"] = 0x33
	m["'"] = 0x34
	m[","] = 0x36
	m["."] = 0x37

	return m[keyCode]
}
