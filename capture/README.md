#Capture networking calls for automatic instrumentation!

GoVector's capture library automatically injects vector clocks onto network reads and writes.
The capture functionality works on the standard [go net libary](https://golang.org/pkg/net/).

#Installing

To install GoVector's capture tool run
```
go install
```
In GoVector's main directory

#Dependencies

GoVector depends on [Dinv](https://bitbucket.org/bestchai/dinv) for static analysis.

#How it works

GoVector capture takes either a go file or directory containing a go package as a command line argument.
The package or file is then translated into its [AST](https://golang.org/pkg/go/ast/) representation.
GoVector traverses the AST searching for read and write calls to go's networking library. When matching calls are identified Capture 
performs an AST rotation and injects wrapper code for the networking call. The wrapper code appends and truncates vector clocks from
network payloads. The set of library wrapping function are in [capture.go](https://github.com/DistributedClocks/GoVector/blob/master/capture/capture.go)
The result of running Capture is an augmented file, or directory. Case specific auto instrumentation's of both Go's RPC, and HTTP are 
also implemented by Capture. Below is an example of a captured write.

###Before
```n, err := conn.Write(buf)``

###After
```n, err := capture.Write(conn.Write,buf)``

#Running


To run Capture on a single file run 
`GoVector -f=filename.go`

To run Capture on a directory containing a single package run
`GoVector -dir=directory`

#Limitations

##Scale
Building an AST in go requires that all referenced files are read in while building the AST. Capture will fail if a single go file
with references to another local file is passed in. In this case the -dir option is suggests. The AST loading library only works on a 
single package at a time. Therefore Capture cannot instrument more than one package at a time.

##Aliasing
Only calls to Go's standard networking library are completely supported. The net.Conn type implements io.ReadWriter, therefore it
can be passed to an encoder or decoder directly. Capture makes no attempt to reason about which encoders or decoders could be
attached to network connections. Encoder wrapped connections will not be instrumented.

