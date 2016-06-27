package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"bufio"
	"fmt"
	"github.com/arcaneiceman/GoVector/capture"
)

const (
	//instrumenter defaults
	defaultFilename  = ""
	defaultDirectory = ""
	defaultPipe = ""
)

var (
	//options for detecting
	directory string
	file      string
	pipe      string

	//options for both
	verbose bool
	debug   bool
	clean   bool

	logger *log.Logger
)

func setFlags() {
	flag.StringVar(&directory, "dir", defaultDirectory, "-dir=directoryName recursivly instruments a directory inplace, original directory is duplicated for safty")
	flag.StringVar(&file, "file", defaultFilename, "-file=filename insturments a file")

	flag.BoolVar(&verbose, "verbose", false, "-verbose logs extensive output")
	flag.BoolVar(&verbose, "v", false, "-v logs extensive output")
	flag.BoolVar(&debug, "debug", false, "-debug adds pedantic level of logging")
	flag.BoolVar(&clean, "c", false, "cleans up old insturmentation")
	flag.Parse()
}

func main() {
	setFlags()

	options := make(map[string]string)
	//set options relevent to all programs
	if verbose {
		logger = log.New(os.Stdout, "logger: ", log.Lshortfile)
	} else {
		var buf bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	}

	if debug {
		options["debug"] = "on"
	}

	if clean {
		options["clean"] = "on"
	}

	//TODO complain about arguments not ment for the instrumenter

	//filechecking //exclusive or with filename and directory
	if file == defaultFilename && directory == defaultDirectory {
		if len(os.Args) == 2 && !verbose{
			file = os.Args[1]
		} else {
			//try to read from pipe
			reader := bufio.NewReader(os.Stdin)
			// Read all data from stdin, processing subsequent reads as chunks.
			data := make([]byte, 100000) // Read 4MB at a time
			n, err := reader.Read(data)
			if err != nil {
				logger.Fatalf("Problems reading from input: %s", err)
			}
			buffer := bytes.NewBuffer(data)
			pipe += buffer.String()[0:n]
		}
	} else if file != defaultFilename && directory != defaultDirectory {
		logger.Fatalf("Speficied filename =%s and directory = %s, use either -file or -dir\n", file, directory)
	}

	if pipe != defaultPipe {
		options["pipe"] = pipe
		source := capture.InsturmentComm(options)
		fmt.Print(source)
		return
	}

	//test if file exists, if so add file option
	if file != defaultFilename {
		exists, err := fileExists(file)
		if !exists {
			a := err.Error()
			print(a)
			logger.Fatalf("Error: : %s\n", err.Error())
		}
		logger.Printf("Documenting %s\n",file)

		options["file"] = file
		//get source
		source := capture.InsturmentComm(options)

		//overwrite file
		file, err := os.OpenFile(file,os.O_RDWR,os.FileMode(0666)) // For read access.
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = file.Truncate(0)
		if err != nil {
			log.Fatal(err)
		}
		logger.Println("Writing over source");
		_, err = file.WriteString(source);
		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO remove test if the directory is valid. If so add to options, else
	// error
	if directory != defaultDirectory {
		valid, err := validDir(directory)
		if !valid {
			logger.Fatalf("Invalid Directory Error: %s\n", err.Error())
		}
		logger.Printf("Documenting Directory :%s\n", directory)
		options["directory"] = directory

		source := capture.InsturmentComm(options)
		fmt.Println(source)
	}


}

func validDir(dir string) (bool, error) {
	//TODO check that dir exists
	//TODO check for existing go args
	/*if len(args) != 3 {
		return false, fmt.Errorf("Directory or package non existant\n")
	}*/
	return true, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
