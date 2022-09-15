#!/bin/bash -e

go install golang.org/x/mobile/cmd/gobind@latest
go install golang.org/x/mobile/cmd/gomobile@latest

go mod download golang.org/x/mobile
go get golang.org/x/mobile/bind

# Force Android API to >= 19 to work around bug with newer SDK:
# https://github.com/golang/go/issues/52470#issuecomment-1203874724
gomobile bind -target=android -androidapi 19 -v .

TARGET="../Android/ultrablue/app/libs"
mkdir -p $TARGET
echo "Copying gomobile.aar to $TARGET"
cp gomobile.aar "$TARGET"
