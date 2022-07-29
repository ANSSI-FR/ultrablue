#!/bin/bash
#
# SPDX-FileCopyrightText: 2022 ANSSI
# SPDX-License-Identifier: Apache-2.0

# Prerequisite check(s) for module.
check() {
    # If the binary(s) requirements are not fulfilled the module can't be installed.
    require_binaries ultrablue-server || return 1
    return 255
}

# Module dependency requirements.
depends() {
    # This module has external dependency on other module(s).
    echo bluetooth tpm2-tss
    # Return 0 to include the dependent module(s) in the initramfs.
    return 0
}

# Install the required file(s) and directories for the module in the initramfs.
install() {
    inst_multiple -o \
        /usr/bin/ultrablue-server \
        "${systemdsystemunitdir}"/ultrablue-server.service

    $SYSTEMCTL -q --root "$initdir" enable ultrablue-server.service
}
