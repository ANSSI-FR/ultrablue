# Ultrablue
Ultrablue (User-friendly Lightweight TPM Remote Attestation over Bluetooth) is a solution to allow individual users to perform boot state attestation with their phone.


It consists of a server, running on a computer, acting as the attester, and a client graphical application, running on a trusted phone, acting as the verifier.

## Installation
To install Ultrablue, please refer to the specific documentation:
**[Linux server](https://github.com/ANSSI-FR/ultrablue/tree/dev/server) / [IOS client](https://github.com/ANSSI-FR/ultrablue/tree/dev/clients/ios) / [Android client](https://github.com/ANSSI-FR/ultrablue/tree/dev/clients/android)**
## Usage
The classical Ultrablue control flow consists of several steps:

### 1. Enrollment
To enroll a phone as a verifier, start the server in enroll mode. This will display a QR code on the terminal. From the phone, run the client app, and tap the **+** icon on the top right corner to show a QR code scanner. On scan, an encrypted Bluetooth Low Energy channel will be established, and the enrollment will run automatically. Upon success, a device card will appear on the home page of the client application.

### 2. Initramfs configuration (optional)
Once enrolled, you have to re-generate your initramfs in order to include the **ultrablue dracut module** in it,
you hence have to install `server/dracut/90ultrablue` in the `/usr/lib/dracut/modules.d/` module directory. You can
then run the following dracut command:

```
dracut --add ultrablue /path/to/initrd --force
```

That's it, you can pass to the attestation part.

### 3. Attestation
If you did the initramfs configuration step, Ultrablue server will run automatically during the boot. Otherwise, manually start the server in attestation mode. Once started, the server will wait for a verifier (phone) to connect.

From the phone, click on the **▶️** icon of the device card. This will run the attestation. When finished, the client application will display the attestation result.

---
The Ultrablue project has been developped at ANSSI ([ssi.gouv.fr](http://ssi.gouv.fr)) by Loïc Buckwell, under the supervision of Nicolas Bouchinet and Gabriel Kerneis.
