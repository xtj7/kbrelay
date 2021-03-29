#!/bin/bash
bluetoothctl untrust $1
bluetoothctl remove $1
bluetoothctl scan on &
echo "Turn your keyboard off and on again within 10 seconds"
sleep 10
bluetoothctl pair $1
bluetoothctl trust $1
bluetoothctl connect $1
kill $(ps aux | grep "sudo bluetoothctl scan on" | awk '{print $2}')
