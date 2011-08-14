/*
   Copyright (c) Utkan Güngördü <utkan@freeconsole.org>

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as
   published by the Free Software Foundation; either version 3 or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details


   You should have received a copy of the GNU General Public
   License along with this program; if not, write to the
   Free Software Foundation, Inc.,
   51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

// Go client for quantum random number generator service at random.irb.hr
package qrand

import (
	"sync"
	"net"
	"fmt"
	"bytes"
 	"encoding/binary"
	"os"
)

const (
	Host = "random.irb.hr"
	Port = "1227"
)

var response = []string {
	"OK",
	"Service was shutting down",
	"Server was/is experiencing internal errors",
	"Service said we have requested some unsupported operation",
	"Service said we sent an ill-formed request packet",
	"Service said we were sending our request too slow",
	"Authentication failed",
	"User quota exceeded",
}
    
var remedy = []string {
	"None",
	"Try again later",
	"Try again later",
	"Upgrade your client software",
	"Upgrade your client software",
	"Check your network connection",
	"Check your login credentials",
	"Try again later, or contact service admin to increase your quota(s)",
}



type QRand struct {
	buffer, b []byte
	buffersize int
	user, pass string
	l sync.Mutex
}

// Read requests len(rand) bytes from QRNG server.
// This function does not user the buffer and
// always creates a new connection.
// Try not to use this function —the function
// you're looking for is probably ReadBytes().
func (q *QRand) Read(rand []byte) (int, os.Error) {
	if len(rand) == 0 { return 0, nil }

	c, err := net.Dial("tcp", net.JoinHostPort(Host,Port))
	if err != nil { return 0, err }
	defer c.Close()

	
	b := bytes.NewBuffer([]byte(""))
	fmt.Fprintf(b, "%c", 0)
	binary.Write(b, binary.BigEndian, uint16(len(q.user) + len(q.pass) + 6))
	fmt.Fprintf(b, "%c%s%c%s", len(q.user), q.user, len(q.pass), q.pass)
	binary.Write(b, binary.BigEndian, uint32(len(rand)))
	_, err = c.Write(b.Bytes())
	if err != nil { return 0, err }

	msg := make([]byte, 6)
	
	_, err = c.Read(msg)
	if err != nil { return 0, err }
	b = bytes.NewBuffer(msg)
	var remedyCode, responseCode uint8
	binary.Read(b, binary.BigEndian, &responseCode)
	binary.Read(b, binary.BigEndian, &remedyCode)
	
	if responseCode != 0 || remedyCode != 0 {
		return 0, os.NewError("qrand: " + response[responseCode] + ": " + remedy[remedyCode])
	}
	
	var available uint32
	binary.Read(b, binary.BigEndian, &available)
	_, err = c.Read(rand[:available])
	return int(available), err
}

// ReadData tries to read len(b) bytes of data into b.
// It returns the number of bytes actually read, which can be less than len(b).
func (q *QRand) ReadBytes(p []byte) (int, os.Error) {
	if len(p) == 0 { return 0, nil }

	q.l.Lock()
	defer q.l.Unlock()

	// We have enough data in the buffer
	if len(q.b) >= len(p) {
		copy(p, q.b[:len(p)])
		q.b = q.b[len(p):]
		return len(p), nil
	}
	
	read := 0
	// First empty the buffer
	copy(p, q.b)
	p = p[len(q.b):]
	read += len(q.b)
	q.b = q.b[:0]

	// If required data is greater than buffer size, directly read into p
	if len(p) > len(q.buffer) {
		n, err := q.Read(p)
		read += n
		if err != nil { return read, err }
		return read, err
	}
	
	// Fill in the buffer, and read from it.
	n, err := q.Read(q.buffer)
	if err != nil { return n, err }
	q.b = q.buffer[:n]
	
	if len(q.b) >= len(p) {
		copy(p, q.b[:len(p)])
		read += len(p)
		q.b = q.b[len(p):]
		return read, nil
	}
	
	// Shouldn't happen normally
	copy(p, q.b)
	read += len(q.b)
	q.b = q.b[:0]
	return read, nil
}

func (q *QRand) readBytes(n int) ([]byte, os.Error) {
	rand := make([]byte, n)
	read, err := q.ReadBytes(rand)
	if err != nil {return rand, err}
	if read != n { return rand, os.NewError(fmt.Sprintf("qrand: Receieved insufficient data; requested: %d, received: %d", n, read)) }
	return rand, nil
}

// Uint8 fetches 8-bit random data and returns it as uint8.
func (q *QRand) Uint8() (uint8, os.Error) {
	data, err := q.readBytes(1)
	return data[0], err
}

// Int8 fetches 8-bit random data and returns it as int8.
func (q *QRand) Int8() (int8, os.Error) {
	data, err := q.readBytes(1)
	return int8(data[0]), err
}

// Uint16 fetches 16-bit random data and returns it as uint16.
func (q *QRand) Uint16() (r uint16, err os.Error) {
	data, err := q.readBytes(2)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

// Int16 fetches 16-bit random data and returns it as int16.
func (q *QRand) Int16() (r int16, err os.Error) {
	data, err := q.readBytes(2)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

// Uint32 fetches 32-bit random data and returns it as uint32.
func (q *QRand) Uint32() (r uint32, err os.Error) {
	data, err := q.readBytes(4)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

// Int32 fetches 32-bit random data and returns it as int32.
func (q *QRand) Int32() (r int32, err os.Error) {
	data, err := q.readBytes(4)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

// Uint64 fetches 64-bit random data and returns it as uint64.
func (q *QRand) Uint64() (r uint64, err os.Error) {
	data, err := q.readBytes(8)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

// Int64 fetches 64-bit random data and returns it as int64.
func (q *QRand) Int64() (r int64, err os.Error) {
	data, err := q.readBytes(8)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

// Float32 fetches 32-bit random data and returns it as a float32 in [0.0,1.0)
func (q *QRand) Float32() (r float32, err os.Error) {
	n, err := q.Int32()
	if err != nil { return 0, err }
	return float32(n)/(1<<32-1-1), err
}

// Float64 fetches 64-bit random data and returns it as a float64 in [0.0,1.0)
func (q *QRand) Float64() (r float64, err os.Error) {
	n, err := q.Int64()
	if err != nil { return 0, err }
	return float64(n)/(1<<64-1-1), err
}

// NewQRand creates a new instances of Quantum Random Bit Generator client.
// The client-side should have a username and password
// from the relevant web-site.
// When host and/or port are empty,
// they are replaced by the default values, Host and Port.
func NewQRand(user, pass string, buffersize int, host, port string) (*QRand, os.Error) {
	if buffersize < 1 { return nil, os.NewError("qrand: buffersize is too small.") }
	if host == "" { host = Host }
	if port == "" { port = Port }
	return &QRand{ user: user, pass: pass, buffer: make([]byte, buffersize) }, nil
}
