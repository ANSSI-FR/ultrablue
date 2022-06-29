# Ultrablue server

## Usage

To compile and the Ultrablue server, enter those commands from the repo base directory:
```
cd server
go build
./ultrablue-server
```

You can use the following flags with the Ultrablue server:
```
--enroll:
	When used, the server will start in enroll mode,
	needed to register a new verifier with the client app.

--loglevel:
	The loglevel flag takes an integer parameter between 0 and 3.
	It indicates the verbosity level of the server.
	0 stands for no log, 2 for maximum output.

--tpmpath:
	The tpmpath takes the path of the TPM to use as parameter
```

⚠️ The server is only Linux compatible for now.

## Behavior

### steps:
When started, the Ultrablue server will do serveral things:

- First, if the enroll flag is set, it will display a QR code to scan with the client app. This qrcode contains the needed data to establish an encrypted [BLE](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy) communication.
- Then it will register a [BLE](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy) service, and characteristics, that implement the attestation protocol we designed, described below, and wait for a verifier to connect and interact with it.
### Prococol:

![Attestation protocol, splitted by characteristics](../protocol/characteristics_protocol.svg)

- **Red box:**  A BLE characteristic.
- **Black arrow:**  A machine local action.
- **Gray arrow:** A message transmitted with a QR code.
- **Blue arrow:** A message transmitted over Bluetooth.
- **Dashed arrow:** An optional message.
- **Yellow highlighted text:** An AES encrypted message
- **Orange highlighted text:** A message encrypted with the EkPub

TODO: Document CBOR encoding, of each message.
