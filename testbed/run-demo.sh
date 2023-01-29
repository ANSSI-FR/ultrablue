#!/bin/bash -e

if [ "$EUID" -ne 0 ]
  then echo "Please run as root."
  exit 1
fi

mkdir -p /tmp/emulated_tpm/ultrablue

swtpm socket \
          --tpmstate dir=/tmp/emulated_tpm/ultrablue \
          --ctrl type=unixio,path=/tmp/emulated_tpm/ultrablue/swtpm-sock \
          --log level=20 --tpm2 --daemon

trap cleanup 1 2 3 6

cleanup()
{
  pkill swtpm
  exit
}

usbflags=()
for d in $(find /sys/class/bluetooth -mindepth 1 -maxdepth 1); do
        usb_device="${d}/device/.."
        usbflags+=(-usb -device usb-host,hostbus=$(cat ${usb_device}/busnum),hostaddr=$(cat ${usb_device}/devnum))
done

if (( ${#usbflags[@]} == 0 )); then
    echo "Couldn't find any USB Bluetooth device. Try: rfkill unblock bluetooth"
fi

IMG="mkosi.output/fedora~36/ultrablue.raw"

qemu-system-x86_64 \
        -machine type=q35,accel=kvm,smm=on -smp 1 -m 1G \
        -object rng-random,filename=/dev/urandom,id=rng0 -device virtio-rng-pci,rng=rng0,id=rng-device0 \
        -cpu max -nographic -nodefaults -serial mon:stdio \
        -drive if=pflash,format=raw,readonly=on,file=OVMF/OVMF_CODE.secboot.fd \
        -global ICH9-LPC.disable_s3=1 -global driver=cfi.pflash01,property=secure,value=on \
        -drive file=OVMF/OVMF_VARS.fd,if=pflash,format=raw \
        -drive "if=none,id=hd,file=${IMG},format=raw" \
        -device virtio-scsi-pci,id=scsi -device scsi-hd,drive=hd,bootindex=1 \
        -chardev socket,id=chrtpm,path=/tmp/emulated_tpm/ultrablue/swtpm-sock -tpmdev emulator,id=tpm0,chardev=chrtpm -device tpm-tis,tpmdev=tpm0 \
        "${usbflags[@]}"

