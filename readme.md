# Bluetooth USB Keyboard Relay

## What is this even?

This application allows you to utilise a computer with dual-role USB controllers (such as a Raspberry Pi Zero or 4) to work as an active USB dongle for bluetooth keyboards. Basically it allows you to connect a blueooth keyboard to another computer that does not have/need bluetooth.

## When is this useful?

My primary use-case is a KVM switch for multiple computers and me wanting to use a bluetooth keyboard without a USB dongle (Apple Magic Keyboard) with it.

If I were to use a regular bluetooth USB receiver, not only would the handover when switching between computers often take several seconds, every now and then it would not work at all on the login screen, requiring the keyboard to be switched off and on again before it would work. Turns out bluetooth devices do not like being constantly switched around nonstop.

Using this tool I can connect the keyboard to the Raspberry Pi 4, have the Pi's USB connect to the KVM keyboard USB port and as the keyboard always stays connected to the Pi, the keyboard will immediately work when switching the KVM between different computers.

If your KVM supports bluetooth out of the box, fantastic. Mine does not, so this was my solution.

## Mapping file

You can create a JSON5 mapping file for your own keyboard (I already provided one for the Apple Magic Keyboard with NumPad as an example) and load it with `-map 