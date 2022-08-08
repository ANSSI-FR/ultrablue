#!/bin/bash

if [ ! -v ANDROID_HOME ]; then
	echo "Setting ANDROID_HOME to ~/Android/Sdk"
	export ANDROID_HOME="/home/lfalkau/Android/Sdk"
fi

if [ ! -v ANDROID_NDK_HOME ]; then
	echo "Setting ANDROID_NDK_HOME to ~/Android/Sdk/ndk-bundle"
	export ANDROID_NDK_HOME="$ANDROID_HOME/ndk-bundle"
fi

go install golang.org/x/mobile/cmd/gobind@latest
go install golang.org/x/mobile/cmd/gomobile@latest

go mod download golang.org/x/mobile
go get golang.org/x/mobile/bind

gomobile bind -target=android -v .

if [ $# -eq 1 ]; then
	echo "Copying gomobile.aac to $1"
	cp gomobile.aar $1
else
	echo "Don't forget to copy gomobile.aac in your android project"
fi
