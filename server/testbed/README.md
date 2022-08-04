# Virtual testbed setup

This testbed allows you to generate a linux distribution image (ie. archlinux, fedora, debian) with ultrablue server
configured in it.

Your machine needs `swtpm` installed and **bluetooth** devices.

The virtual testbed generation is made by the `mkosi` tool you hence have to install it. A Makefile is available in
order to generate the virtual testbed, feel free to appen a `--distribution <distroname>` option to the `mkosi`
commands to specify the distribution you want to generate, which is by default the host one. Beware that `mkosi`
needs root privileges in order to work and will write the `mkosi` and `mkosi.output` cache directories as root
owned.

## Testing ultrablue

Once your distro image is generated, you can boot it using the `make run` command, then unlock the disk using the
passphrase written in the `mkosi.passphrase` file.

You can now enroll your `ultrablue client` using the `ultrablue --enroll` command. Once enrolled, you have to
re-generate your initramfs in order to include the ultrablue dracut module in it, it can be done using the
following dracut command:

```
dracut --add ultrablue /path/to/initrd --force
```

You can now reboot the testbed and use your `ultrablue client` once asked in order to attest your machine boot.
