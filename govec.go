package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// Command Line arguments

var (
	logType      = flag.String("log_type", "", "Type of the log that needs to be generated (Shiviz or TSViz)")
	logDirectory = flag.String("log_dir", "", "Input directory which has individual node logs")
	outputFile   = flag.String("outfile", "", "The file in which the log will be written")
)

func parse_args() {
	flag.Parse()
	if *logType == "" || *logDirectory == "" || *outputFile == "" {
		fmt.Println("Usage: GoVector --log_type [Shiviz | TSViz] --log_dir [directory] --outfile [output_file] ")
		os.Exit(1)
	}
}

func get_regex(logType string) string {
	t := strings.ToLower(logType)
	if t == "shiviz" {
		return "(?<host>\\S*) (?<clock>{.*})\\n(?<event>.*)"
	} else if t == "tsviz" {
		return "(?<timestamp>\\d+) (?<host>\\S*) (?<clock>{.*})\\n(?<event>.*)"
	}

	return ""
}

func write_log(logDirectory string, outputFile string, logType string) {
	files, err := ioutil.ReadDir(logDirectory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outf, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer outf.Close()

	regex := get_regex(logType)
	outf.Write([]byte(regex + "\n\n"))

	for _, f := range files {
		fname := f.Name()
		if strings.HasSuffix(fname, "Log.txt") {
			filepath := path.Join(logDirectory, fname)
			content, err := ioutil.ReadFile(filepath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			outf.Write(content)
		}
	}
}

func main() {
	parse_args()
	write_log(*logDirectory, *outputFile, *logType)
}
