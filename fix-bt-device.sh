#!/bin/bash
sudo bluetoothctl untrust $1
sudo bluetoothctl remove $1
sudo bluetoothctl scan on &
echo "Turn your keyboard off and on again within 10 seconds"
sleep 10
sudo bluetoothctl pair $1
sudo bluetoothctl trust $1
sudo bluetoothctl connect $1
sudo kill $(ps aux | grep "sudo bluetoothctl scan on" | awk '{print $2}')
