## About

Go client for quantum random number generator service at http://random.irb.hr

## Installation

You can use goinstall to install the library

     goinstall github.com/salviati/go-qrand

or clone & build manually

     git clone git://github.com/salviati/go-qrand
     cd go-qrand
     make install     

Then you can import the package and start using it

     import qrand "github.com/salviati/go-qrand"
     ...
     q, err := qrand.NewQRand(user, pass, cachesize, qrand.Host, qrand.Port)
     if err != nil { ... }
     rnd := make([]byte, 16)
     nread, err := q.ReadBytes(rnd)

See `example/example.go` for a full demo.

## Documentation

You can use `godoc` to and browse the documentation locally. Or you can browse online at http://gopkgdoc.appspot.com/pkg/github.com/salviati/go-qrand
