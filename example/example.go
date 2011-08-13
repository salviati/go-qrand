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


package main

import (
	"flag"
	"qrand"
	"fmt"
	"log"
)

var (
	user = flag.String("u", "", "Username for QRNG server")
	pass = flag.String("p", "", "Password for QRNG server")
	cachesize = flag.Int("c", 1024, "Cache size")
	host = flag.String("host", qrand.Host, "Host name")
	port = flag.String("port", qrand.Port, "Port")
)

func main() {
	flag.Parse()
	
 	q, err := qrand.NewQRand(*user, *pass, *cachesize, *host, *port)
 	if err != nil { log.Fatal(err) }

	i8, err := q.Int8()
	if err != nil { log.Fatal(err) }
	fmt.Println(i8)
	
	i16, err := q.Int16()
	if err != nil { log.Fatal(err) }
	fmt.Println(i16)
	
	i32, err := q.Int32()
	if err != nil { log.Fatal(err) }
	fmt.Println(i32)
	
	i64, err := q.Int64()
	if err != nil { log.Fatal(err) }
	fmt.Println(i64)
	
	buf := make([]byte, 128)
	_, err = q.ReadBytes(buf)
	if err != nil { log.Fatal(err) }
	fmt.Println(buf)
}