// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

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

/*
	recvBLEPacket reads a BLE packet in the @req buffer, and appends it
	in the final @out message.
	If *@size is zero (meaning it is the first packet's message),
	recvBLEPacket first reads the size, prefix and update the size.
	Note that if a protocol message is fixed length, and the sender don't
	put the size at the message start, *@size must be set manually before
	the first recvBLEPacket call.

	Returns true if the whole message has been received, false otherwise.
*/
func recvBLEPacket(out *[]byte, size *int, req ble.Request) bool {
	var offset int

	if *size == 0 && len(req.Data()) >= 4 {
		*size = int(binary.LittleEndian.Uint32(req.Data()[0:4]))
		offset = 4
	}
	*out = append(*out, req.Data()[offset:]...)
	return (len(*out) == *size)
}
