# Ultrablue, a remote attestation server for your phone

Ultrablue (User-friendly Lightweight TPM Remote Attestation over Bluetooth) is
a solution to allow individual users to perform boot state attestation with
their phone.
It consists in a server, running on a computer, acting as the attester, and a
graphical client application, running on a trusted phone, acting as the
verifier.

A typical use-case is to verify the integrity of your bootchain before
unlocking your computer, to prevent offline attacks on an unattended laptop. It
can also serve as a debugging tool for secure boot issues after firmware
upgrades or a second factor for disk encryption.

Once installed, the classical Ultrablue control flow consists of several steps:

1. Enrollment
2. Boot-time integration in the initramfs (optional)
3. Attestation

## 0. Installation

Using Ultrablue requires a Linux computer that you want to attest the boot state of,
featuring a TPM and a Bluetooth interface (typically a laptop), as well as mobile phone.

* On the computer, you need to build and install the **[Linux server](server).
* On the mobile phone, you need to install either the [IOS client](clients/ios)
  or the [Android client](clients/Android)**.


## 1. Enrollment

To enroll a phone as a verifier, start the server in enroll mode:

```
sudo ultrablue-server -enroll -pcr-extend
```

This will display a QR code on the terminal. From the phone, run the client
app, and tap the **+** icon on the top right corner to show a QR code scanner.
On scanning, a Bluetooth Low Energy channel will be established, and the
enrollment will run automatically. Upon success, a device card will appear on
the home page of the client application.

**Ultrablue** can extend your **TPM2 PCR9** using a randomly generated value at
enroll time. This is usefull if you want to, eg., bind your disk encryption to
**TPM2 sealing**. In that case, **ultrablue** will extend back the **PCR9**
register during boot-time if the attestation is successfull and trusted.
PCR-extension is configured at enroll time (the flag has no effect in
attestation mode):

```
sudo ultrablue-server -enroll -pcr-extend
```

### 2. Boot-time integration using Dracut (optional)

If you want ultrablue to execute as part of your boot flow, you have to
re-generate your initramfs to bundle it in. We make this easier by providing
Dracut and systemd integration. See also [the provided VM
scripts](server/testbed) which provide an even easier way to test this.

First, install `server/dracut/90ultrablue` in the `/usr/lib/dracut/modules.d/` module directory,

You can then run the following dracut command:

```bash
dracut --add ultrablue /path/to/initrd --force
```

If you used the `--pcr-extend` option during the enrollment phase, you'll need
to add the **crypt** dracut module, and to copy`server/dracut/90crypttab.conf` in `/etc/dracut.conf.d` (to work around [dracut bug #751640](https://bugzilla.redhat.com/show_bug.cgi?id=751640#c18)):

```bash
cp server/dracut/90crypttab.conf /etc/dracut.conf.d
dracut --add "crypt ultrablue" /path/to/initrd --force
```

Note that those options are not persistent and **ultrablue** will be removed
from your initramfs on its next generation. See the dracut.conf(5) man page for
persistent configuration.

## 3. Attestation

If you did the initramfs configuration step, Ultrablue server will run
automatically during the boot. Otherwise, manually start the server in
attestation mode:
```bash
sudo ultrablue-server
```

Once started, the server will wait for a verifier (phone) to connect. From the
phone, click on the **▶️** icon of the device card. This will run the
attestation. When finished, the client application will display the attestation
result.

## 4. Disk decryption based on remote attestation

The main goal of running ultrablue at boot time is to use it for disk decryption.
An example of how to do this is provided and documented in the [server
testbed](server/testbed).

## Contact

The Ultrablue project has been developped at ANSSI
([ssi.gouv.fr](http://ssi.gouv.fr)) by Loïc Buckwell, under the supervision of
Nicolas Bouchinet and Gabriel Kerneis.

If you have any question, reach out via a Github issue, or directly to Gabriel
Kerneis <gabriel.kerneis@ssi.gouv.fr>.
