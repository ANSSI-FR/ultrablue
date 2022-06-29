// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"

	"github.com/go-ble/ble"
	"github.com/sirupsen/logrus"
)

type OpKind int

const (
	Idle = iota
	Read
	Write
)

/*
	State struct is used to keep informations about the current
	attestation state, between the attester (server) and a
	verifier (client).
*/
type State struct {

	/*
		Buf is the buffer in which the messages that are exchanged
		between the client and the server are stored while they
		are processed.
		As the messages are chunked in several packets, we need
		to keep track of additional data, Offset and Msglen.
	*/
	Buf []byte

	/*
		Indicates wether we're reading, writing or idling.
	*/
	operation OpKind

	/*
		Offset is the number of bytes from a message that the server
		already sent to the client, if we are in a write operation,
		-1 otherwise.
	*/
	Offset int

	/*
		Msglen is the size of the message the server expects to
		receive from the client if we're in a read operation,
		-1 otherwise.
	*/
	Msglen int

	/*
		messageCh is a channel used to transmit messages between the
		linear ultrablueProtocol function and ultrablueChr callbacks.
		It is a bidirectional channel.
		In case of failure during the BLE message exchange, the characteristic
		will close this channel, close the connection, and the ultrablueProtocol
		function will return.
	*/
	ch chan []byte
}

// Key type to get/set the state value for
// a context.
type key int

/*
	getConnectionState returns the attestation state
	for the given connection. If there's no
	value, it sets it with a newly created state, and
	returns it.

	This is useful to keep a separate state for each
	connection, and avoid leaks.
*/
func getConnectionState(conn ble.Conn) *State {
	var connCtx = conn.Context()
	var stateKey key

	if connCtx.Value(stateKey) == nil {
		s := &State{
			ch: make(chan []byte),
		}
		s.reset()
		ctx := context.WithValue(connCtx, stateKey, s)
		conn.SetContext(ctx)
		connCtx = ctx
		// Start the attestation protocol that runs in a goroutine
		// and reads/receives messages through the channel.
		go ultrablueProtocol(s.ch)
	}
	return connCtx.Value(stateKey).(*State)
}

func (s *State) isComplete() bool {
	if s.operation == Read && len(s.Buf) == s.Msglen {
		return true
	}
	if s.operation == Write && len(s.Buf) == s.Offset {
		return true
	}
	return false
}

func (s *State) reset() {
	s.Buf = nil
	s.Msglen = -1
	s.Offset = -1
	s.operation = Idle
}

/*
	check is an internal State method that
	asserts the state is valid.
	Exits in case of error.
*/
func (s *State) check() {
	var isValid bool

	switch s.operation {
	case Idle:
		isValid = len(s.Buf) == 0 && s.Offset == -1 && s.Msglen == -1
	case Read:
		isValid = s.Msglen >= 0 && s.Offset == -1
	case Write:
		isValid = s.Offset >= 0 && s.Msglen == -1
	}

	if isValid == false {
		logrus.Fatalf("Invalid state: operation=%d, Buf length=%d, Offset=%d, Msglen=%d", s.operation, len(s.Buf), s.Offset, s.Msglen)
	}
}

/*
	StartOperation puts the connection in a "busy" state, meaning
	a message is being sent or received. If another call to StartOperation
	is made before EndOperation, it will raise an error.
	In other words, StartOperation and EndOperation calls must match one to one.
*/
func (s *State) StartOperation(kind OpKind) error {
	s.check()
	if kind == Idle {
		return errors.New("Cannot start an Idle operation.")
	}
	if s.operation != Idle {
		return errors.New("An operation is already in progress")
	}
	s.operation = kind
	switch kind {
	case Read:
		s.Msglen = 0
	case Write:
		s.Offset = 0
	}
	s.check()
	return nil
}

/*
	EndOperation puts the connection back in the Idle state. This indicates
	that the client is now able to send/receive a message. Internally, it
	clears the message buffer and helper variables.
	The function is closely related to StartOperation func, see above.
	It also checks that the contract of the message has been fullfilled,
	which implies that either:
		- The number of sent bytes matches the message size.
		- The number of received bytes matches the message size prefix.
	if those condition aren't satisfied, an error is raised.
*/
func (s *State) EndOperation() error {
	s.check()
	if s.operation == Idle {
		return errors.New("There is no operation in progress")
	}

	if !s.isComplete() {
		switch s.operation {
		case Read:
			return errors.New("the read operation hasn't been completed")
		case Write:
			return errors.New("the write operation hasn't been completed")
		}
	}
	s.reset()
	s.check()
	return nil
}
