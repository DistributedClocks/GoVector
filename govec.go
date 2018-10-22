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
    log_type = flag.String("log_type", "", "Type of the log that needs to be generated (Shiviz or TSViz)")
    log_directory = flag.String("log_dir", "", "Input directory which has individual node logs")
    output_file = flag.String("outfile", "", "The file in which the log will be written")
)

func parse_args() {
    flag.Parse()
    if *log_type == "" || *log_directory == "" || *output_file == "" {
        fmt.Println("Usage: GoVector --log_type [Shiviz | TSViz] --log_dir [directory] --outfile [output_file] ")
        os.Exit(1)
    }
}

func get_regex(log_type string) string {
    t := strings.ToLower(log_type)
    if t == "shiviz" {
        return "(?<host>\\S*) (?<clock>{.*})\\n(?<event>.*)"
    } else if t == "tsviz" {
        return "(?<timestamp>\\d+) (?<host>\\S*) (?<clock>{.*})\\n(?<event>.*)"
    }

    return ""
}

func write_log(log_directory string, output_file string, log_type string) {
    files, err := ioutil.ReadDir(log_directory)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    outf, err := os.Create(output_file)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer outf.Close()

    regex := get_regex(log_type)
    outf.Write([]byte(regex + "\n\n"))

    for _, f := range files {
        fname := f.Name()
        if strings.HasSuffix(fname, "Log.txt") {
            filepath := path.Join(log_directory, fname)
            content, err :=  ioutil.ReadFile(filepath)
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
    write_log(*log_directory, *output_file, *log_type)
}
