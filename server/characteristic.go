// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

/*
	The functions in this file implement an abstraction layer
	to send and receive messages on a BLE channel.

	The interface with the rest of the application is a
	bidirectionnal go channel carrying byte slices.
	Each send/receive operation is made by writing/reading
	on the go channel, and is blocking.
	Acknowledgment of successful write operation is notified
	with a nil message written back on the go channel.

	On the BLE channel, messages are chunked in packets of
	MTU max size, because there's no native concept of packet
	fragmentation in BLE.
	The first packet of each message starts with the size of the
	message, encoded on four bytes, in little endian, followed
	by the raw message bytes.
	There's no prefix in the following chunks.
*/

package main

import (
	"errors"

	"encoding/binary"
	"github.com/go-ble/ble"
	"github.com/sirupsen/logrus"
)

const DEFAULT_MTU = 20

/*
	terminateConnection is an error handling helper:
	In case of error in the transport layer, we need to
	close the channel if the characteristic is expected
	to write on it, to signal the error to the reader on the
	other side. If the characteristic is instead expected to read from
	the channel, it must not be closed, thus nil must be given as parameter.
	The connection is closed regardless, disconnecting the client,
	but not terminating the program. This means that the client
	can reconnect later to retry the attestation.
*/
func terminateConnection(conn ble.Conn, ch chan []byte) {
	if ch != nil {
		close(ch)
	}
	conn.Close()
}

/*
	sendBLEPacket performs a partial write of @msg of at most @mtu bytes,
	starting at offset @off, to the @rsp BLE channel.
	It updates @off to point to the first unwritten byte.
	If @off is equal to 0, the size of the full message is written as
	prefix in little endian.

	If the message is not fully sent during a call to sendBLEPacket,
	the remote device should read the characteristic again, to get
	the remaining bytes.
*/
func sendBLEPacket(off *int, msg []byte, mtu int, rsp ble.ResponseWriter) error {
	var packet []byte
	var copied, msgoff int

	if *off >= len(msg) {
		return nil
	}

	if len(msg) >= 1<<32 {
		logrus.Fatal("Message too big:", len(msg))
	}

	packet = make([]byte, mtu)
	if *off == 0 {
		binary.LittleEndian.PutUint32(packet[:], uint32(len(msg)))
		msgoff = 4
	}
	copied = copy(packet[msgoff:mtu], msg[*off:])
	_, err := rsp.Write(packet[:copied+msgoff])
	if err != nil {
		return err
	}
	*off += copied
	return nil
}

/*
	recvBLEPacket reads a BLE packet in the @req buffer, and appends it
	in the final @out message.
	If *@size is zero (meaning it is the first packet's message),
	recvBLEPacket first reads the size, prefix and update the size.
	Note that if a protocol message is fixed length, and the sender don't
	put the size at the message start, *@size must be set manually before
	the first recvBLEPacket call.
*/
func recvBLEPacket(out *[]byte, size *int, req ble.Request) error {
	var offset int

	if *size == 0 && len(req.Data()) >= 4 {
		*size = int(binary.LittleEndian.Uint32(req.Data()[0:4]))
		offset = 4
	}
	if *size < 0 {
		return errors.New("Invalid packet size prefix")
	}
	*out = append(*out, req.Data()[offset:]...)
	return nil
}

/*
	UltrablueChr is the only characteristic exposed by the ultrablue server.
	The client server interactions are made through the HandleRead and HandleWrite
	callbacks.
	In order to abstract chunking, those callbacks will store / chunk / rebuild the
	messages internally, and send / receive full messages in an internal channel
	when some message is fully available.
	This makes the server able to interact with the client easily through channels.
*/
func UltrablueChr(mtu int) *ble.Characteristic {
	chr := ble.NewCharacteristic(ultrablueChrUUID)

	if mtu < 20 || mtu > 500 {
		mtu = DEFAULT_MTU
	}

	// HandleRead is a callback triggered when a client reads on the characteristic,
	// which corresponds to the server sending data (Write operation)
	chr.HandleRead(ble.ReadHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		logrus.Tracef("%s - HandleRead", ultrablueChrUUID.String())

		var state = getConnectionState(req.Conn())

		if state.operation != Write {
			if err := state.StartOperation(Write); err != nil {
				terminateConnection(req.Conn(), state.ch)
				return
			}
			var ok bool
			state.Buf, ok = <-state.ch
			if !ok {
				// The channel has already been closed.
				terminateConnection(req.Conn(), nil)
				return
			}
			logrus.Tracef("New message to send:\nlen: %d, preview: %v\n", len(state.Buf), state.Buf)
		}
		err := sendBLEPacket(&state.Offset, state.Buf, mtu, rsp)
		if err != nil {
			logrus.Error(err)
			terminateConnection(req.Conn(), state.ch)
			return
		}
		logrus.Tracef("Sent %d/%d bytes - %d%%", state.Offset, len(state.Buf), int(100.0*state.Offset/len(state.Buf)))
		if state.isComplete() {
			if err := state.EndOperation(); err != nil {
				logrus.Error(err)
				terminateConnection(req.Conn(), state.ch)
				return
			}
			// Sending nil to the channel notifies the caller that the operation has ended.
			state.ch <- nil
		}
	}))

	// HandleWrite is a callback triggered when a client writes on the characteristic,
	// which corresponds to the server expecting data (Read operation)
	chr.HandleWrite(ble.WriteHandlerFunc(func(req ble.Request, rsp ble.ResponseWriter) {
		logrus.Tracef("%s - HandleWrite", ultrablueChrUUID.String())

		var state = getConnectionState(req.Conn())

		if state.operation != Read {
			if err := state.StartOperation(Read); err != nil {
				logrus.Error(err)
				terminateConnection(req.Conn(), state.ch)
				return
			}
		}
		err := recvBLEPacket(&state.Buf, &state.Msglen, req)
		if err != nil {
			terminateConnection(req.Conn(), state.ch)
			return
		}
		if state.isComplete() {
			logrus.Tracef("Received a message:\nlen:%d, preview: %v\n", state.Msglen, state.Buf)
			buf := state.Buf
			if err := state.EndOperation(); err != nil {
				logrus.Error(err)
				terminateConnection(req.Conn(), state.ch)
				return
			}
			state.ch <- buf
		}
	}))

	return chr
}
