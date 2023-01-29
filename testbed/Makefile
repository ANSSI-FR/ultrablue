# SPDX-FileCopyrightText: 2022 ANSSI
# SPDX-License-Identifier: Apache-2.0
#
hostbus := $(shell lsusb | grep Bluetooth | sed -r 's/.*Bus 0*([0-9]+){1}.*/\1/')
hostaddr := $(shell lsusb | grep Bluetooth | sed -r 's/.*Device 0*([0-9]+){1}.*/\1/')

# Allow building with `make DISTRIBUTION=debian`
# `man mkosi` for a list of distributions
DISTRIBUTION=

MKOSI_FLAGS=
ifeq ($(DISTRIBUTION),)
else
	MKOSI_FLAGS += --distribution=$(DISTRIBUTION)
endif

# A few useful aliases - by default, build and run.
quick: build run
full: scratch-build run
debug: debug-build run
build: cache-build

scratch-build: install
	sudo mkosi $(MKOSI_FLAGS) -ff

cache-build: install
	sudo mkosi $(MKOSI_FLAGS) --incremental -f

debug-build: install
	sudo mkosi $(MKOSI_FLAGS)  --kernel-command-line="systemd.log_level=debug systemd.log_target=console" --force

install:
	mkdir -p mkosi.extra/usr/lib/dracut/modules.d
	cp -r ../server/dracut/90ultrablue mkosi.extra/usr/lib/dracut/modules.d/
	mkdir -p mkosi.extra/etc/dracut.conf.d
	cp ../server/dracut/90crypttab.conf mkosi.extra/etc/dracut.conf.d
	mkdir -p mkosi.extra/usr/lib/systemd/system/
	cp ../server/unit/ultrablue-server.service mkosi.extra/usr/lib/systemd/system/
	mkdir -p mkosi.extra/etc
	cp ressources/crypttab mkosi.extra/etc/crypttab

run:
	sudo mkosi $(MKOSI_FLAGS) qemu \
	    -usb -device usb-host,hostbus=${hostbus},hostaddr=${hostaddr}

