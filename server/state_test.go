// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"github.com/go-ble/ble"
)

/*
	All the following pain, only to implement the Conn interface
	of the go-ble package, which doesn't provide a test type.
*/

type FakeConn struct {
	context context.Context
}

func (fc *FakeConn) Context() context.Context {
	return fc.context
}

func (fc *FakeConn) SetContext(ctx context.Context) {
	fc.context = ctx
}

func (fc *FakeConn) LocalAddr() ble.Addr {
	return ble.NewAddr("42:42:42:42:42:42")
}

func (fc *FakeConn) RemoteAddr() ble.Addr {
	return ble.NewAddr("42:42:42:42:42:42")
}

func (fc *FakeConn) RxMTU() int {
	return 0
}

func (fc *FakeConn) SetRxMTU(mtu int) {

}

func (fc *FakeConn) TxMTU() int {
	return 0
}

func (fc *FakeConn) SetTxMTU(mtu int) {

}

func (fc *FakeConn) ReadRSSI() int {
	return 0
}

func (fc *FakeConn) Close() error {
	return nil
}

func (fc *FakeConn) Write(p []byte) (int, error) {
	return len(p), nil
}

func (fc *FakeConn) Read(p []byte) (int, error) {
	return len(p), nil
}

func (fc *FakeConn) Disconnected() <-chan struct{} {
	return make(chan struct{})
}

/*
	Tests that the getConnectionState function well
	sets the stateKey value for the connection
	context on the first call
*/
func TestGetConnectionState(t *testing.T) {
	var conn = FakeConn{
		context: context.Background(),
	}

	var stateKey key
	if conn.Context().Value(stateKey) != nil {
		t.Errorf("Context has value for stateKey key: %+v", conn.Context())
	}
	_ = getConnectionState(&conn)
	if conn.Context().Value(stateKey) == nil {
		t.Errorf("Context value for stateKey key is nil: %+v", conn.Context())
	}
}

func TestStartOperation_NormalCase(t *testing.T) {
	var s = State{}
	s.reset()

	err := s.StartOperation(Read)
	if err != nil {
		t.Error("StartOperation failed, whereas no operation is in progress")
	}
}

func TestStartOperation_AnotherOperationInProgress(t *testing.T) {
	var s = State{
		operation: Read,
		Msglen:    0,
		Offset:    -1,
	}

	err := s.StartOperation(Write)
	if err == nil {
		t.Error("StartOperation succeeded whereas an operation is running")
	}
}

func TestEndOperation_WholeMessageWritten(t *testing.T) {
	var s = State{
		operation: Write,
		Buf:       make([]byte, 8),
		Offset:    8,
		Msglen:    -1,
	}
	err := s.EndOperation()
	if err != nil {
		t.Errorf("EndOperation failed, whereas an operation completed successfully: %+v", s)
	}
	if s.Buf != nil {
		t.Errorf("The buffer hasn't been reset correctly: %+v", s)
	}
	if s.Offset != -1 {
		t.Errorf("The offset hasn't been reset correctly: %+v", s)
	}
}

func TestEndOperation_BufBiggerThanMsglen(t *testing.T) {
	var s = State{
		operation: Read,
		Buf:       make([]byte, 16),
		Msglen:    12,
		Offset:    -1,
	}
	err := s.EndOperation()
	if err == nil {
		t.Errorf("EndOperation succeeded whereas the running operation wasn't valid: %+v", s)
	}
}

func TestEndOperation_ReadingNotComplete(t *testing.T) {
	var s = State{
		operation: Read,
		Buf:       make([]byte, 16),
		Msglen:    20,
		Offset:    -1,
	}
	err := s.EndOperation()
	if err == nil {
		t.Errorf("EndOperation succeeded whereas the running operation wasn't complete: %+v", s)
	}
}
