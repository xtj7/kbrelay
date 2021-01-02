#!/bin/bash
stty -echo
sudo go run kbrelay.go
stty echo