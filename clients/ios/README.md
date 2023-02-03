# Ultrablue IOS client

The IOS client has not been merged yet, see [PR#14](https://github.com/ANSSI-FR/ultrablue/pull/14).
=======

Welcome file

# Ultrablue IOS client

This directory contains an implementation of the Ultrablue client for IOS.

## Getting started

⚠️ A MacOS system is required to build the IOS client. ⚠️

1. Download Xcode for your platform:
   https://apps.apple.com/us/app/xcode/id497799835?mt=12
2. Launch Xcode and import the `ultrablue` project.
3. Configure a device to run the app:
4. **The IOS client depends on the go-mobile framework.**
   You need to build it first: see instructions in [clients/go-mobile](../go-mobile/README.md).
5. Import the go-mobile framework into the Xcode project:
   https://www.simpleswiftguide.com/how-to-add-xcframework-to-xcode-project/
7. You are now ready to build and run the app:
   https://developer.android.com/studio/run

## How to test the app

Ultrablue is tested manually using an iPhone 13 mini.
There is no way to fully test the client in the IOS emulator because the
emulator does not support bluetooth.

### Enrollment:
Start the server in enroll mode:
```
sudo server/ultrablue_server --enroll
```
This should display a QR Code.

Start the client app on your phone (from Xcode), and hit the ➕ button on the top right corner. Scan the QR code on the computer, and let things run. When the enrollment ends, a new card should appear for that computer.

### Attestation:

Start the server:
```
sudo server/ultrablue_server
```
Start the client app on your phone (from Xcode), and hit the ▶️ button on the right of your computer card. 
This will run an attestation. In case of attestation failure, follow the instructions from the IOS application.

### Use Ultrablue for disk decryption:

If you want to use Ultrablue for disk decryption, use the `pcr-extend` flag at enroll time. 
You will then need to embeed Ultrablue in your initramfs, and base the disk decryption on the PCR 9.

[Here](https://github.com/ANSSI-FR/ultrablue/tree/dev/server/testbed) are some information on how to configure such a test environment.

## Troubleshooting

* If your build fails with an error message about gomobile, make sure you have built [the go-mobile library](../go-mobile/README.md).
