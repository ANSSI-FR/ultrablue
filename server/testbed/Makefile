# SPDX-FileCopyrightText: 2022 ANSSI
# SPDX-License-Identifier: Apache-2.0
#
hostbus := $(shell lsusb | grep Bluetooth | sed -r 's/([^0-9]*([0-9]*)){1}.*/\2/')
hostaddr := $(shell lsusb | grep Bluetooth | sed -r 's/([^0-9]*([0-9]*)){2}.*/\2/')

all: install scratch-build run
quick: install cache-build run
debug: install debug-build run

scratch-build:
	sudo mkosi -ff

cache-build:
	sudo mkosi --incremental -f

debug-build:
	sudo mkosi --kernel-command-line="systemd.log_level=debug systemd.log_target=console" --force

install:
	mkdir -p mkosi.extra/usr/lib/dracut/modules.d
	cp -r ../dracut/90ultrablue mkosi.extra/usr/lib/dracut/modules.d/
	mkdir -p mkosi.extra/usr/lib/systemd/system/
	cp ../unit/ultrablue-server.service mkosi.extra/usr/lib/systemd/system/
	cp ../unit/bluetooth.service mkosi.extra/usr/lib/systemd/system/
	mkdir -p mkosi.extra/etc
	cp ressources/crypttab mkosi.extra/crypttab

run:
	mkdir -p /tmp/emulated_tpm/ultrablue
	
	swtpm socket \
	          --tpmstate dir=/tmp/emulated_tpm/ultrablue \
	          --ctrl type=unixio,path=/tmp/emulated_tpm/ultrablue/swtpm-sock \
	          --log level=20 --tpm2 --daemon
	
	sudo mkosi qemu \
	    -chardev socket,id=chrtpm,path=/tmp/emulated_tpm/ultrablue/swtpm-sock \
	    -tpmdev emulator,id=tpm0,chardev=chrtpm \
	    -device tpm-tis,tpmdev=tpm0 \
	    -usb -device usb-host,hostbus=${hostbus},hostaddr=${hostaddr}

