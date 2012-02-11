## About

Go client for quantum random number generator service at http://random.irb.hr

## Installation

You can use go tool to install the library

     go get github.com/salviati/go-qrand
 
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
