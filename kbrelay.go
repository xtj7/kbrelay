package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type EnabledKey struct {
	enabled bool
	altKey  string
}

var debugEnabled *bool
var enabledKeys map[string]EnabledKey
var forwardKeys bool

func main() {
	debugEnabled = flag.Bool("debug", false, "Enables / disables debug mode")
	flag.Parse()

	setupCloseHandler()
	go setupKeyboardHandlers()
	dummyInputHandler()
}

func dummyInputHandler() {
	// Throw away all input data
	for {
		var password string
		fmt.Println("\033[8m") // Hide input
		fmt.Scan(&password)
		fmt.Println("\033[28m") // Show input
	}
}

func setupKeyboardHandlers() {
	forwardKeys = false

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

	enabledKeys := make(map[string]EnabledKey)

	// open output
	f, err := os.OpenFile("/dev/hidg0", os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatal(err)
	}

	// range of events
	for e := range events {
		//logrus.Println("Received event", e.Type, e.Code, e.Time)
		switch e.Type {
		// EvKey is used to describe state changes of keyboards, buttons, or other key-like devices.
		// check the input_event.go for more events
		case keylogger.EvKey:
			var keyString = e.KeyString()
			var altKeyString = fmt.Sprintf("KEY_%d", e.Code)
			if keyString == "" {
				keyString = altKeyString
			}

			// if the state of key is pressed
			if e.KeyPress() {
				if *debugEnabled == true {
					logrus.Println("[event] press key ", keyString, e.Code, keyCodeToScanCode(keyString, altKeyString))
				}
				enabledKeys[keyString] = EnabledKey{
					enabled: true,
					altKey:  altKeyString,
				}
			}

			// if the state of key is released
			if e.KeyRelease() {
				if *debugEnabled == true {
					logrus.Println("[event] release key ", keyString, e.Code, keyCodeToScanCode(keyString, altKeyString))
				}
				enabledKeys[keyString] = EnabledKey{
					enabled: false,
					altKey:  altKeyString,
				}
			}

			if enabledKeys["KEY_188"].enabled && enabledKeys["KEY_189"].enabled {
				forwardKeys = !forwardKeys
				logrus.Println("Changed forwardKeys to", forwardKeys)
				if enabledKeys["ESC"].enabled {
					os.Exit(0)
				}
			}

			if forwardKeys {
				sendKeys(enabledKeys, f)
			}

			break
		}
	}
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
	}()
}

type SendKeyType struct {
	key    string
	altKey string
}

func sendKeys(enabledKeys map[string]EnabledKey, f *os.File) {
	var keyList []SendKeyType
	for keyName, data := range enabledKeys {
		if data.enabled && !isModifierKey(keyName) {
			keyList = append(keyList, SendKeyType{
				key:    keyName,
				altKey: data.altKey,
			})
		}
	}

	var buf bytes.Buffer
	buf.WriteRune(rune(getModifierCode(enabledKeys)))
	if len(keyList) > 0 {
		buf.WriteRune(rune(0))
		for i := 0; i < 6; i++ {
			if i < len(keyList) {
				buf.WriteRune(rune(keyCodeToScanCode(keyList[i].key, keyList[i].altKey)))
			} else {
				buf.WriteRune(rune(0))
			}
		}
	} else {
		// Release all keys
		for i := 0; i < 7; i++ {
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

func getModifierCode(enabledKeys map[string]EnabledKey) int {
	// iterate over all keys, check for modifier keys, return bit sum of all modifier keys
	var returnModifierCode int = 0
	for keyName, data := range enabledKeys {
		if data.enabled && isModifierKey(keyName) {
			modifierCode, _ := getModifierCodeForKey(keyName)
			returnModifierCode += modifierCode
		}
	}
	return returnModifierCode
}

func getModifierCodeForKey(keyName string) (int, bool) {
	m := make(map[string]int)
	m["KEY_126"] = 0
	m["R_CTRL"] = 1
	m["R_SHIFT"] = 2
	m["R_ALT"] = 4
	m["KEY_125"] = 8
	m["L_CTRL"] = 16
	m["L_SHIFT"] = 32
	m["L_ALT"] = 64

	val, ok := m[keyName]
	return val, ok
}

func keyCodeToScanCode(keyCode string, altKeyCode string) int {
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

	m["ENTER"] = 0x28
	m["ESC"] = 0x29
	m["BS"] = 0x2a
	m["TAB"] = 0x2b
	m["SPACE"] = 0x2c
	m["-"] = 0x2d
	m["{"] = 0x2e
	m["["] = 0x2f
	m["]"] = 0x30
	m["\\"] = 0x31
	m["~"] = 0x32
	m[";"] = 0x33
	m["'"] = 0x34
	m["KEY_86"] = 0x35 // grave
	m[","] = 0x36
	m["."] = 0x37
	m["/"] = 0x38
	m["CAPS_LOCK"] = 0x39

	m["F1"] = 0x3a
	m["F2"] = 0x3b
	m["F3"] = 0x3c
	m["F4"] = 0x3d
	m["F5"] = 0x3e
	m["F6"] = 0x3f
	m["F7"] = 0x40
	m["F8"] = 0x41
	m["F9"] = 0x42
	m["F10"] = 0x43
	m["F11"] = 0x44
	m["F12"] = 0x45

	m[""] = 0x46 // Print Screen
	m[""] = 0x47 // Scroll Lock
	m[""] = 0x48 // Pause
	m[""] = 0x49 // Insert
	m["Home"] = 0x4a
	m["PgUp"] = 0x4b
	m["Del"] = 0x4c
	m["End"] = 0x4d
	m["PgDn"] = 0x4e
	m["Right"] = 0x4f
	m["Left"] = 0x50
	m["Down"] = 0x51
	m["Up"] = 0x52

	//m["KEY_117"] = 0x65 // KP Equals

	m["NUM_LOCK"] = 0x53
	m["KEY_98"] = 0x54  // KP Slash
	m["*"] = 0x55       // KP *
	m["KEY_74"] = 0x56  // KP -
	m["KEY_78"] = 0x57  // KP +
	m["R_ENTER"] = 0x58 // KP ENTER
	m[""] = 0x59        // KP 1
	m[""] = 0x5a        // KP 2
	m[""] = 0x5b        // KP 3
	m[""] = 0x5c        // KP 4
	m[""] = 0x5d        // KP 5
	m[""] = 0x5e        // KP 6
	m[""] = 0x5f        // KP 7
	m[""] = 0x60        // KP 8
	m[""] = 0x61        // KP 9
	m["KEY_82"] = 0x62  // KP 0
	m["KEY_83"] = 0x63  // KP Dot/Delete

	m["`"] = 0x64 // Key next to left shift

	m["KEY_113"] = 0x7f // Mute
	m["KEY_115"] = 0x80 // Volume Up
	m["KEY_114"] = 0x81 // Volume Down
	m["KEY_164"] = 0xe8 // Media: PlayPause
	m["KEY_165"] = 0xea // Media: PreviousSong
	m["KEY_163"] = 0xeb // Media: NextSong
	m["KEY_161"] = 0xec // Media: EjectCD

	if val, ok := m[altKeyCode]; ok {
		return val
	} else {
		return m[keyCode]
	}
}
