# Virtual testbed setup

This testbed allows you to generate a linux distribution image (eg. archlinux,
fedora, debian) with an embedded and configured ultrablue server.

## Quickstart

```bash
make build
# The passphrase to unlock the disk is in specified in mkosi.passphrase.
make run
```

The VM should start and log you in as root. From there:

```bash
# Check that PCR 9 is all-zero
tpm2_pcrread
# Run this command (only once!) and add a new device from Ultrablue phone application
ultrablue-server -enroll -pcr-extend
# Check that PCR 9 is *not* all-zero
tpm2_pcrread
# Lock the disk based on the PCR value
systemd-cryptenroll --tpm2-device=auto --tpm2-pcrs=9 /dev/vda2
# Rebuild the initramfs to enable Ultrablue on boot
dracut --add "crypt ultrablue" --force $(find /efi -name initrd)
# Reboot and click "run" on your machine in the mobile app when Ultrablue starts
reboot
```

If everythings works as expected, you should not need to input the passphrase upon reboot.

## Building the image

```bash
make build
```

Your machine needs `swtpm` installed and a **bluetooth** device. The virtual
testbed generation is made by the `mkosi` tool, that needs to be installed on
your machine. 

The distribution defaults to the host one; you can build a different one with
eg. `make DISTRIBUTION=debian`. Beware that `mkosi` needs root privileges in
order to work and will write the `mkosi` and `mkosi.output` cache directories
as root owned.

The build script will automatically build and embbed ultrablue-server, but you
need to build the client app of your choice and install it on your phone
separately.

## Remote attestation

Once your distro image is generated, you can boot it using the `make run` command, then unlock the disk using the
passphrase written in the `mkosi.passphrase` file.

You can now enroll your `ultrablue client`:
```bash
ultrablue-server -enroll
```

Then, you can test remote attestation:
```bash
ultrablue-server
```
and press the "run" button on your mobile client to start the attestation.


## Disk decryption based on remote attestation

If you want to bind your disk encryption to **TPM2 sealing** and **ultrablue's
remote attestation**, you'll then have to use the `--pcr-extend` option during
enrollment.

Beware that if you enrolled without `--pcr-extend`, as in the previous section,
you'll have to enroll again. The cleanest way to do so is to remove the machine
record from your Ultrablue mobile application, reboot the VM, and enroll once
more:
```bash
ultrablue-server -enroll -pcr-extend
```

Bind your disk encryption to the **TPM2 PCR9**:
```bash
systemd-cryptenroll --tpm2-device=auto --tpm2-pcrs=9 /dev/vda2
```

Re-generate your initramfs in order to include the ultrablue and crypt dracut modules:
```bash
# Dracut has issues finding the right initrd on Debian, provide a hint.
dracut --add "crypt ultrablue" --force $(find /efi -name initrd)
```

You can now reboot the testbed and use your `ultrablue client` once asked in
order to attest your machine boot.

