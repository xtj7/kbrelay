# Bluetooth USB Keyboard Relay

---

## What is this even?

This application allows you to utilise a computer with dual-role USB controllers (such as a Raspberry Pi Zero or 4) to
work as an active USB dongle for bluetooth keyboards. Basically it allows you to connect a blueooth keyboard to another
computer that does not have/need bluetooth.

## When is this useful?

My primary use-case is a KVM switch for multiple computers and me wanting to use a bluetooth keyboard without a USB
dongle (Apple Magic Keyboard) with it.

If I were to use a regular bluetooth USB receiver, not only would the handover when switching between computers often
take several seconds, every now and then it would not work at all on the login screen, requiring the keyboard to be
switched off and on again before it would work. Turns out bluetooth devices do not like being constantly switched around
nonstop.

Using this tool I can connect the keyboard to the Raspberry Pi 4, have the Pi's USB connect to the KVM keyboard USB port
and as the keyboard always stays connected to the Pi, the keyboard will immediately work when switching the KVM between
different computers.

If your KVM supports bluetooth out of the box, fantastic. Mine does not, so this was my solution.

## Prerequisites

You need to have go and python3 installed on your system. I think that's it.

To install `python` simply run:

```
sudo apt update
sudo apt install python3 idle3
```

To install `go` follow this guide:
https://www.jeremymorgan.com/tutorials/raspberry-pi/install-go-raspberry-pi/

You *could* use apt for this as well, but unfortunately the packages are horribly outdated at the time of writing. If
this changes, let me know and I will update the guide.

## Installation

1. Checkout this repo, ideally locally in `/home/pi/kbrelay` (makes the rest of the guide easier to follow)
2. Switch to the repo directory and run `sudo ./install.sh`
3. Run `sudo ./start.sh`

Ideally the help page will show up and you should be able to start typing, no longer seeing the input on your screen.
The inputs should be forwarded via USB.

To exit, by default press F18+F19+ESC. If you screwed up the host keys configuration, you can always switch to another
TTY and kill the process.

## Mapping file

You can create a JSON5 mapping file for your own keyboard (I already provided one for the Apple Magic Keyboard with
NumPad as an example) and load it with `-map /home/pi/kbrelay/maps/apple-magic-keyboard-numpad.json5`

---

# FAQs / tips

## Help

The most important commands are printed right when you start the application via `./start.sh`

## Host keys

By default the host keys are mapped to F18 + F19 for the Apple keyboard.

That means in order to stop forwarding keys, you can just hit F18 + F19 + F.

In order to exit the application, you can hit F18 + F19 + ESC.

## How do I automatically run this app on startup?

You will likely want to run this script as soon as the raspi starts, so that you can use your keyboard immediately once
it boots up. Add the following code to your `.bashrc`:

```kbd_mode -u
sudo systemctl start bluetooth
sudo bluetoothctl scan on &
sleep 5
sudo ~/kbrelay/install.sh
~/kbrelay/start.sh
```

This assumes this application is located in your home folder (by default `/home/pi`) in a folder called `kbrelay`.

## Can I run this script with the GUI enabled?

Short answer: yes you can, but it is not recommended, you will likely do all sorts of things you do not intend to.

Long answer: As we can receive all keyboard events, but we cannot intercept them (and thus stop them from going to the
system), I recommend you to not run the GUI on your raspi. Not intercepting them means: all keys you press will still be
registered by the system, so you might open up menus, might end up googling your password by accident or similar things.
It is just a bad idea. See here how to switch to console mode: https://itsfoss.com/raspberry-pi-gui-boot/

## How do I connect my keyboard via bluetooth from the shell?

If your keyboard is not connected already, run `sudo bluetoothctl scan on`, wait for the keyboard to be found (you might
have to turn it off and on again so that it will be in discovery mode - ensure it is not currently connected to another
computer) and once you see `NEW` popping up, remember / write down the ID (like 11:22:33:44:55:66).

Then you can run `sudo ./connect-bt-keyboard.sh 11:22:33:44:55:66`, just ensure to replace the ID with the actual device
ID.

---

# Known issues

## The code quality... is not great

This was more of a prototyping application meant to be rewritten. But since I posted a photo of my setup on reddit, I
got a lot of questions about how I resolved the problem of wirelessly connecting the Apple keyboard to multiple
computers through a KVM. Several people were interested, so I created a readme and published this first. Feel free to
contribute, but a rewrite will be coming... eventually :)

## Mapping not 100% complete

We have to send scancodes to the virtual keyboard in order for it to "press" a key. Linux conveniently maps scancodes to
keycodes, so they are always identical for applications to use. There is a tool called `evtest` that lets you press keys
and get the scancodes (MSC_SCAN event) additionally to the keycode, so that you can map them out. However, sadly evtest
does not return scancodes for all keys (like the media play/pause, prev/next keys), so that you need to try out
scancodes blindly until you hit the right one.

I am going to implement a "learning" mode soon, that should finding these keys easier, as you will then be able to "
step" through keys until it does the action you intend for it to do.

## Not automatically reconnecting if the keyboard disconnects

Once the keyboard disconnects, the application terminates. You can simply start it again once the keyboard is
reconnected. This is on my todo list and should be a relatively small fix.

## It automatically connects to the first keyboard found

As I only have one keyboard connected to the raspi, this was not an issue for me. But if you happen to also have a wired
keyboard connected (or another device that pretends to be a keyboard), this may be an issue for you. If you need this,
submit an issue in Github and I will create a flag so you can pass the device ID on startup.

## Python is only required for the setup script

I have used a setup script by [Kosci](https://koscis.wordpress.com/2018/11/23/raspberry-pi-as-usb-gadget-part-1/) to
create the USB gadget, because it worked and I was just prototyping. However, I will likely rewrite this into a bash
version so that we can get rid of this (then) unnecessary dependency. 