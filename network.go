package main

import (
	"encoding/binary"
	"io"
	"net"
	"syscall"
	"time"
)

func send(conn io.Writer, buf []byte) {
	l := len(buf)
	mid := (l + 1) / 2
	n, err := conn.Write(buf[0:mid])
	time.Sleep(200 * time.Millisecond)
	println("Sent bytes: ", n)
	if err != nil {
		panic(err)
	}
	if n != mid {
		panic("didn't send all bytes")
	}

	n, err = conn.Write(buf[mid:])

	println("Sent bytes: ", n)
	if err != nil {
		panic(err)
	}
	if n != l-mid {
		panic("didn't send all bytes")
	}
}

func readRecord(reader io.Reader) Record {
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	if err != nil {
		panic(err)
	}
	if n != 5 {
		panic("didn't read 5 bytes")
	}
	length := binary.BigEndian.Uint16(buf[3:])
	contents := read(int(length), reader)
	return concatenate(buf, contents)
}

func read(length int, reader io.Reader) []byte {
	var buf []byte
	for len(buf) != length {
		buf = append(buf, readUpto(length-len(buf), reader)...)
	}
	return buf
}

func readUpto(length int, reader io.Reader) []byte {
	buf := make([]byte, length)
	n, err := reader.Read(buf)
	if err != nil {
		panic(err)
	}
	return buf[:n]
}

func connect() Session {
	receive_buffer_size := 2
	dialer := &net.Dialer{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			if err := conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, receive_buffer_size)
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, receive_buffer_size)
				println("SetsockoptInt: ", operr)
			}); err != nil {
				return err
			}
			return operr
		},
	}

	conn, err := dialer.Dial("tcp", "localhost:10000")
	if err != nil {
		panic(err)
	}

	return Session{Conn: conn}
}
