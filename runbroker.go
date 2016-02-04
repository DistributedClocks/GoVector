package main

import (
    "fmt"
    "net"
    "net/http"
    "flag"
    "./broker"
)
/*
    This program runs a Vector Broker from the host machine. It initializes a 
    broker with the given log path (logpath - optional), publisher port 
    (pubport), and subscriber port (subport). It also prints the available 
    addresses for connection by publishers and subscribers.
    
    Publishers are intended to connect via the govec library and GoPublishers.
    Subscribers connect via a Websocket connection and receive messages sent
    to the broker.
*/

var broker brokervec.VectorBroker

//Printing out the various ways the server can be reached by the clients
func printClientConnInfo(pubport string, subport string) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("Oops: " + err.Error())
        return
    }

    fmt.Println("Chat clients can connect at the following addresses:\n")

    for _, a := range addrs {
        if a.String() != "0.0.0.0" {
            fmt.Println("http://" + a.String() + "\n")
        }
    }
    fmt.Println("Publishers connect to port:    " + pubport + " and")
    fmt.Println("Subscribers connect to port:    " + subport + "\n")
}

func printUsage() {
    fmt.Println("Usage is: go run ./runbroker (-logpath logpath) -pubport pubport -subport subport")
}

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./broker/static/"+r.URL.Path)
}

//#############MAIN FUNCTION and INITIALIZATIONS

// Setup command line parameters for the broker.
var logpath = flag.String("logpath", "", "The log file path, leave blank for no log file.")
var pubport = flag.String("pubport", "", "The port that publishers should connect to.")
var subport = flag.String("subport", "", "The port that subscribers should connect to.")


func main() {
    flag.Parse()
    // If no port was provided for publisher or subscribers then print usage and exit.
    if *pubport == "" || *subport == "" {
        printUsage()
        return
    }

    // Print connection information.
    printClientConnInfo(*pubport, *subport)

    // Setup html files to be served on the default http server.
    http.HandleFunc("/", staticFiles)

    // Initialize the broker with the commandline parameters
    broker.Init(*logpath, *pubport, *subport)
}



