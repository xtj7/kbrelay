#!/bin/bash
bluetoothctl pair $1
bluetoothctl trust $1
bluetoothctl connect $1
