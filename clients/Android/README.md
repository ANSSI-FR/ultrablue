# Ultrablue Android client

This directory contains an implementation of the Ultrablue client for Android.

## Getting started

1. Download Android Studio for your platform:
   https://developer.android.com/studio Ultrablue is developed using Linux,
   other platforms have not been tested.
2. Install Android Studio for your platform:
   https://developer.android.com/studio/install
3. Launch Android Studio and import the `ultrablue` project.
4. Configure a device to run the app:
   https://developer.android.com/studio/run/device
   Under Linux, pay special attention to the part about udev permissions.
5. You are now ready to build and run the app:
   https://developer.android.com/studio/run

## How to test the app

Ultrablue is tested manually using a Pixel 4a.
There is no way to fully test the client in the Android emulator because the
emulator does not support bluetooth.

Start the server in enroll mode:

```
$ sudo server/ultrablue_server --enroll
```

This should display a QR Code.

Start the client app on your phone (from Android Studio), and follow the
instructions displayed by the server to scan the QR Code.

## Troubleshooting

General Android Studio tips, not specific to Ultrablue:

* Under Linux, if your build fails, make sure [the tmp directory has exec
  permissions](https://github.com/xerial/sqlite-jdbc/issues/97#issuecomment-220855060).
* If your emulator fails to start under Wayland, check that you have `xcb`
  listed as a fallback in `$QT_QPA_PLATFORM`. Start the emulator manually in a
  terminal to get useful error messages: `~/Android/Sdk/emulator/emulator @Pixel_4a_API_30`.

