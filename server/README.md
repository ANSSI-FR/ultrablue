# Ultrablue server
## Installation & Usage
Use the following commands to clone, compile and run the Ultrablue server:
```
git clone git@github.com:ANSSI-FR/ultrablue # If you didn't clone it yet
cd ultrablue/server
go build
./ultrablue-server

You may have to run the server as root to access bluetooth and TPM devices.

```
You can use the following flags with the Ultrablue server:
```
--enroll:
	When used, the server will start in enroll mode,
	needed to register a new verifier with the client app.
	Otherwise, the server will start in attestation mode.

--loglevel:
	The loglevel flag takes an integer parameter between 0 and 3.
	It indicates the verbosity level of the server.
	0 stands for no log, 2 for maximum output.

--mtu:
	Sets the MTU (Maximum transmission Unit) size for the BLE packets.
	Must be between 20 and 500 to be effective.
```
---
⚠️ The server is only Linux compatible for now.
