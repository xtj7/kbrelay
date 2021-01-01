package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/sirupsen/logrus"
	"github.com/rolldever/go-json5"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type EnabledKey struct {
	enabled bool
	altKey  string
}

type KbMapData struct {
	keys map[string]int
	modifiers map[string]int
}

var debugEnabled *bool
var enabledKeys map[string]EnabledKey
var forwardKeys bool
var mapData KbMapData

func main() {
	debugEnabled = flag.Bool("debug", false, "Enables / disables debug mode")
	flag.Parse()

	loadData()
	setupCloseHandler()
	go setupKeyboardHandlers()
	dummyInputHandler()
}

func loadData() {
	mapData = loadKbMap("./maps/apple-magic-keyboard-numpad.json5")
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

			if enabledKeys["KEY_187"].enabled && enabledKeys["KEY_189"].enabled {
				loadData()
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
	val, ok := mapData.modifiers[keyName]
	return val, ok
}

func keyCodeToScanCode(keyCode string, altKeyCode string) int {
	if val, ok := mapData.keys[altKeyCode]; ok {
		return val
	} else {
		return mapData.keys[keyCode]
	}
}

func loadKbMap(fileName string) KbMapData {
	absPath, _ := filepath.Abs(fileName)
	jsonFile, err := os.Open(absPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	fmt.Printf("Loaded config file %v\n", absPath)

	var mapData KbMapData
	if err := json5.Unmarshal(byteValue, &mapData); err != nil {
		panic(err)
	}

	return mapData
}
