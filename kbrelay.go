package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/MarinX/keylogger"
	"github.com/rolldever/go-json5"
	"github.com/sirupsen/logrus"
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
	HostKeys []string `json:"hostKeys"`
}

type SendKeyType struct {
	key    string
	altKey string
}

var keyboardFound = false
var debugEnabled *bool
var mapFile *string
var enabledKeys map[string]EnabledKey
var forwardKeys bool
var mapData KbMapData
var kbOutputFile *os.File

func main() {
	handleFlags()
	loadData()
	setupCloseHandler()
	printHelp()
	go setupKeyboardHandlers()
	dummyInputHandler() // needs to be last as it is blocking
}

func handleFlags() {
	path, err := os.Executable()
	if err != nil {
		logrus.Println(err)
	}

	debugEnabled = flag.Bool("debug", false, "Enables / disables debug mode")
	mapFile = flag.String("map", filepath.Dir(path) + "/maps/apple-magic-keyboard-numpad.json5", "Path to map file")
	flag.Parse()
}

func loadData() {
	mapData = loadKbMap(*mapFile)
}

func setupCloseHandler() {
	// Make sure not to close the application with Ctrl+C
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if !keyboardFound {
			os.Exit(0)
		}
		fmt.Println("\r- Ctrl+C pressed in Terminal, ignored. Use HOST KEYS + ESC to quit.")
	}()
}

func printHelp() {
	fmt.Println("### How to use ###")
	fmt.Println("HOST KEYS + ESC\tquit")
	fmt.Println("HOST KEYS + F\ttoggle forwardKeys")
	fmt.Println("HOST KEYS + R\treload keys map")
	fmt.Println("HOST KEYS + S\tsave keys map")
}

func setupKeyboardHandlers() {
	forwardKeys = false
	enabledKeys = make(map[string]EnabledKey)

	// find keyboard device, does not require root permission
	keyboard := keylogger.FindKeyboardDevice()
	if keyboard != "" {
		logrus.Println("Found a keyboard at", keyboard)
		keyboardFound = true
	} else {
		logrus.Println("No keyboard found, aborting")
		os.Exit(1)
	}

	// init keylogger with keyboard
	k, err := keylogger.New(keyboard)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	defer k.Close()

	events := k.Read()

	// open output
	kbOutputFile, err = os.OpenFile("/dev/hidg0", os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
	defer kbOutputFile.Close()

	// everything seems great, now enable key forwarding
	forwardKeys = true

	// range of events
	for e := range events {
		handleKeyEvent(e)
	}

	// channel was closed
	os.Exit(1)
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

		// react to special host key shortcuts or forward keys (if enabled)
		if hostKeysPressed() {
			if enabledKeys["F"].enabled {
				forwardKeys = !forwardKeys
				logrus.Println("Changed forwardKeys to", forwardKeys)
			} else if enabledKeys["R"].enabled {
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

func dummyInputHandler() {
	// Throw away all input data
	for {
		var password string
		fmt.Println("\033[8m") // Hide input
		fmt.Scan(&password)
		fmt.Println("\033[28m") // Show input
	}
}

func hostKeysPressed() bool {
	if len(mapData.HostKeys) == 0 {
		logrus.Errorln("No HostKeys defined in map file, exiting")
		os.Exit(0)
	}
	allKeysPressed := true
	for _, keyString := range mapData.HostKeys {
		if !enabledKeys[keyString].enabled {
			allKeysPressed = false
		}
	}
	return allKeysPressed
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
