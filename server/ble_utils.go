package main

import (
	"encoding/binary"

	"github.com/go-ble/ble"
)

const DEFAULT_MTU = 20

/*
	sendBLEPacket performs a partial write of @msg of at most @mtu bytes,
	starting at offset @off, to the @rsp BLE channel.
	It updates @off to point to the first unwritten byte.
	If @off is equal to 0, the size of the full message is written as
	prefix in little endian.

	Returns true if the whole message has been sent, false otherwise.
	In the latter case, the remote device should read the characteristic
	again, to get the remaining bytes.
*/
func sendBLEPacket(off *int, msg []byte, mtu int, rsp ble.ResponseWriter) (bool, error) {
	var packet []byte
	var copied, msgoff int

	if *off >= len(msg) {
		return true, nil
	}
	if mtu < 20 || mtu > 500 {
		mtu = DEFAULT_MTU
	}
	packet = make([]byte, mtu)
	if *off == 0 {
		binary.LittleEndian.PutUint32(packet[:], uint32(len(msg)))
		msgoff = 4
	}
	copied = copy(packet[msgoff:mtu], msg[*off:])
	_, err := rsp.Write(packet[:copied+msgoff])
	if err != nil {
		return false, err
	}
	*off += copied
	return (*off == len(msg)), nil
}
