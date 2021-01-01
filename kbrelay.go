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
	Keys map[string]int `json:"keys"`
	Modifiers map[string]int `json:"modifiers"`
}

var debugEnabled *bool
var mapFile *string
var enabledKeys map[string]EnabledKey
var forwardKeys bool
var mapData KbMapData
var kbOutputFile *os.File

func main() {
	debugEnabled = flag.Bool("debug", false, "Enables / disables debug mode")
	mapFile = flag.String("map", "./maps/apple-magic-keyboard-numpad.json5", "Path to map file")
	flag.Parse()

	loadData()
	setupCloseHandler()
	go setupKeyboardHandlers()
	dummyInputHandler()
}

func loadData() {
	mapData = loadKbMap(*mapFile)
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

	// open output
	kbOutputFile, err = os.OpenFile("/dev/hidg0", os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatal(err)
	}
	defer kbOutputFile.Close()

	// range of events
	for e := range events {
		handleKeyEvent(e)
	}
}

func handleKeyEvent(e keylogger.InputEvent) {
	switch e.Type {
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
			if enabledKeys["F"].enabled {
				forwardKeys = !forwardKeys
				logrus.Println("Changed forwardKeys to", forwardKeys)
			} else if enabledKeys["L"].enabled {
				loadData()
			} else if enabledKeys["S"].enabled {
				logrus.Println("Save data")
			} else if enabledKeys["ESC"].enabled {
				os.Exit(0)
			}
		} else if forwardKeys {
			sendKeys(enabledKeys)
		}

		break
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

func sendKeys(enabledKeys map[string]EnabledKey) {
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
	buf.WriteTo(kbOutputFile)
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
	val, ok := mapData.Modifiers[keyName]
	return val, ok
}

func keyCodeToScanCode(keyCode string, altKeyCode string) int {
	if val, ok := mapData.Keys[altKeyCode]; ok {
		return val
	} else {
		return mapData.Keys[keyCode]
	}
}

func loadKbMap(fileName string) KbMapData {
	absPath, _ := filepath.Abs(fileName)
	jsonFile, err := os.Open(absPath)
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	fmt.Printf("Loaded config file %v\n", absPath)

	var mapData KbMapData
	if err := json5.Unmarshal(byteValue, &mapData); err != nil {
		panic(err)
	}

	return mapData
}
