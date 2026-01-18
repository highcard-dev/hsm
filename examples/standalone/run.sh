#!/bin/bash

docker build -t hsm ../.. 

# You can just use the hsm (or hsm.exe) binary, instead of $HSM
HSM="docker run -it --rm -v $PWD:/data -v $PWD/config:/home/hsm/.config hsm"

echo "Logging in..."
$HSM login

echo "Downloading..."
$HSM download 
