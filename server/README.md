# Ultrablue server

```
go build
./ultrablue-server
```

You may have to run the server as root to access bluetooth and TPM devices.

## Usage

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

--pcr-extend:
	Extends the 9th PCR with the verifier secret on attestation success.
```

## Testing

```
# GOTMPDIR is only necessary if /tmp is set as noexec.
GOTMPDIR=$XDG_RUNTIME_DIR go test
```

There is also a [testbed to do end-to-end testing](testbed/).
## Configuration files

Ultrablue-server itself has no configuration file.

Sample integration files for systemd and Dracut are provided in the `unit/` and
`dracut/` directories. See [the testbed VM](testbed/) for example usage of those.


---
⚠️ The server is only Linux compatible for now.
