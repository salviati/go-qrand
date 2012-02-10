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
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

const (
	Host = "random.irb.hr"
	Port = "1227"
)

var response = []string{
	"OK",
	"Service was shutting down",
	"Server was/is experiencing internal errors",
	"Service said we have requested some unsupported operation",
	"Service said we sent an ill-formed request packet",
	"Service said we were sending our request too slow",
	"Authentication failed",
	"User quota exceeded",
}

var remedy = []string{
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
	buf        *bufio.Reader
	user, pass string
	sync.Mutex
}

// Read requests len(rand) bytes from QRNG server.
// This function does not user the buffer and
// always creates a new connection.
// Try not to use this function —the function
// you're looking for is probably ReadBytes().
func (q *QRand) Read(rand []byte) (int, error) {
	if len(rand) == 0 {
		return 0, nil
	}

	c, err := net.Dial("tcp", net.JoinHostPort(Host, Port))
	if err != nil {
		return 0, err
	}
	defer c.Close()

	b := bytes.NewBuffer([]byte(""))
	fmt.Fprintf(b, "%c", 0)
	binary.Write(b, binary.BigEndian, uint16(len(q.user)+len(q.pass)+6))
	fmt.Fprintf(b, "%c%s%c%s", len(q.user), q.user, len(q.pass), q.pass)
	binary.Write(b, binary.BigEndian, uint32(len(rand)))
	_, err = c.Write(b.Bytes())
	if err != nil {
		return 0, err
	}

	msg := make([]byte, 6)

	_, err = c.Read(msg)
	if err != nil {
		return 0, err
	}
	b = bytes.NewBuffer(msg)
	var remedyCode, responseCode uint8
	binary.Read(b, binary.BigEndian, &responseCode)
	binary.Read(b, binary.BigEndian, &remedyCode)

	if responseCode != 0 || remedyCode != 0 {
		var resp, rem string

		if int(responseCode) < len(response) {
			resp = response[responseCode]
		} else {
			resp = fmt.Sprint("Unknown response code ", responseCode)
		}

		if int(remedyCode) < len(remedy) {
			rem = remedy[remedyCode]
		} else {
			resp = fmt.Sprint("Unknown remedy code ", remedyCode)
		}

		return 0, errors.New("qrand: " + resp + ": " + rem)
	}

	var available uint32
	binary.Read(b, binary.BigEndian, &available)
	_, err = c.Read(rand[:available])
	return int(available), err
}

// ReadBytes tries to read len(b) bytes of data into b.
// It returns the number of bytes actually read, which can be less than len(b).
// An error is returned if fewer bytes are read.
func (q *QRand) ReadBytes(p []byte) (int, error) {
	q.Lock()
	defer q.Unlock()
	return io.ReadFull(q.buf, p)
}

func (q *QRand) readInto(v interface{}) error {
	n := 0

	switch t := v.(type) {
	case *uint8, *int8:
		n = 1
	case *uint16, *int16:
		n = 2
	case *uint32, *int32, *float32:
		n = 4
	case *uint64, *int64, *float64:
		n = 8
	default:
		return errors.New("qrand.readInto: unexpected data type")
	}

	rand := make([]byte, n)
	if _, err := q.ReadBytes(rand); err != nil {
		return err
	}

	return binary.Read(bytes.NewBuffer(rand), binary.BigEndian, v)
}

// Uint8 fetches 8-bit random data and returns it as uint8.
func (q *QRand) Uint8() (r uint8, err error) {
	err = q.readInto(&r)
	return
}

// Int8 fetches 8-bit random data and returns it as int8.
func (q *QRand) Int8() (r int8, err error) {
	err = q.readInto(&r)
	return
}

// Uint16 fetches 16-bit random data and returns it as uint16.
func (q *QRand) Uint16() (r uint16, err error) {
	err = q.readInto(&r)
	return
}

// Int16 fetches 16-bit random data and returns it as int16.
func (q *QRand) Int16() (r int16, err error) {
	err = q.readInto(&r)
	return
}

// Uint32 fetches 32-bit random data and returns it as uint32.
func (q *QRand) Uint32() (r uint32, err error) {
	err = q.readInto(&r)
	return
}

// Int32 fetches 32-bit random data and returns it as int32.
func (q *QRand) Int32() (r int32, err error) {
	err = q.readInto(&r)
	return
}

// Uint64 fetches 64-bit random data and returns it as uint64.
func (q *QRand) Uint64() (r uint64, err error) {
	err = q.readInto(&r)
	return
}

// Int64 fetches 64-bit random data and returns it as int64.
func (q *QRand) Int64() (r int64, err error) {
	err = q.readInto(&r)
	return
}

// Float32 fetches 32-bit random data and returns it as a float32 in [0.0,1.0)
func (q *QRand) Float32() (float32, error) {
	var r uint32
	err := q.readInto(&r)
	return float32(r) / (1 << 32), err
}

// Float64 fetches 64-bit random data and returns it as a float64 in [0.0,1.0)
func (q *QRand) Float64() (float64, error) {
	var r uint64
	err := q.readInto(&r)
	return float64(r) / (1 << 64), err
}

// NewQRand creates a new instances of Quantum Random Bit Generator client.
// The client-side should have a username and password
// from the relevant web-site.
// When host and/or port are empty,
// they are replaced by the default values, Host and Port.
func NewQRand(user, pass string, buffersize int, host, port string) (*QRand, error) {
	if buffersize < 1 {
		return nil, errors.New("qrand: buffersize is too small.")
	}
	if host == "" {
		host = Host
	}
	if port == "" {
		port = Port
	}

	q := &QRand{user: user, pass: pass}
	buf, err := bufio.NewReaderSize(q, buffersize)
	if err != nil {
		return nil, err
	}
	q.buf = buf
	return q, nil
}
