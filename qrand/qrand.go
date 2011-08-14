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
	"bufio"
)

const (
	Host = "random.irb.hr"
	Port = "1227"
	CacheSizeMin = 8 // One call to bytes() should handle all cases.
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
	user, pass string
	buf *bufio.Reader
	l sync.Mutex
}

// Requests len(rand) bytes from QRNG server.
// This function does not user the buffer and
// always creates a new connection.
// Try not to use this function ---the function
// you're looking for is probably ReadBytes().
func (q *QRand) Read(rand []byte) (int, os.Error) {
	q.l.Lock() // Prevent double dials
	defer q.l.Unlock()

	c, err := net.Dial("tcp", net.JoinHostPort(Host,Port))
	if err != nil { return 0, err }
	defer c.Close()
	
	b := bytes.NewBuffer([]byte(""))
	fmt.Fprintf(b, "%c", 0)
	binary.Write(b, binary.BigEndian, uint16(len(q.user) + len(q.pass) + 6))
	fmt.Fprintf(b, "%c%s%c%s", len(q.user), q.user, len(q.pass), q.pass)
	binary.Write(b, binary.BigEndian, len(rand))
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
		return 0, os.NewError(response[responseCode] + ": " + remedy[remedyCode])
	}
	
	var available uint32
	binary.Read(b, binary.BigEndian, &available)
	_, err = c.Read(rand[:available])
	return int(available), err
}

func (q *QRand) ReadBytes(b []byte) (int, os.Error) {
	return q.buf.Read(b)
}

func (q *QRand) readBytes(n int) ([]byte, os.Error) {
	rand := make([]byte, n)
	read, err := q.ReadBytes(rand)
	if err != nil {return rand, err}
	if read != n { return rand, os.NewError("qrand: Receieved insufficient data.") }
	return rand, nil
}

func (q *QRand) Uint8() (uint8, os.Error) {
	data, err := q.readBytes(1)
	return data[0], err
}

func (q *QRand) Int8() (int8, os.Error) {
	data, err := q.readBytes(1)
	return int8(data[0]), err
}

func (q *QRand) Uint16() (r uint16, err os.Error) {
	data, err := q.readBytes(2)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

func (q *QRand) Int16() (r int16, err os.Error) {
	data, err := q.readBytes(2)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

func (q *QRand) Uint32() (r uint32, err os.Error) {
	data, err := q.readBytes(4)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

func (q *QRand) Int32() (r int32, err os.Error) {
	data, err := q.readBytes(4)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

func (q *QRand) Uint64() (r uint64, err os.Error) {
	data, err := q.readBytes(8)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

func (q *QRand) Int64() (r int64, err os.Error) {
	data, err := q.readBytes(8)
	if err != nil { return 0, err }
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &r)
	return r, err
}

func (q *QRand) Float32() (r float32, err os.Error) {
	n, err := q.Int32()
	if err != nil { return 0, err }
	return float32(n)/(1<<32-1-1), err
}

func (q *QRand) Float64() (r float64, err os.Error) {
	n, err := q.Int64()
	if err != nil { return 0, err }
	return float64(n)/(1<<64-1-1), err
}

func NewQRand(user, pass string, cachesize int, host, port string) (*QRand, os.Error) {
	if cachesize < CacheSizeMin { cachesize = CacheSizeMin }
	if host == "" { host = Host }
	if port == "" { port = Port }
	q := &QRand{ user: user, pass: pass }
	buf, err := bufio.NewReaderSize(q, cachesize)
	if err != nil { return nil, err }
	q.buf = buf
	return q, nil
}
