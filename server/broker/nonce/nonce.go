// Author: Tyler Egeto, tyleregeto.com
// Free to use, modify, and redistribute.

// A simple nonce implmentation. Nonce, (number used once), is a
// simple way of adding extra security to a web application, particularly
// preventing replay attacks, semantic URL attacks, and also verify the
// source of the request. For more information see these resources:
// a) http://tyleregeto.com/a-guide-to-nonce
// b) http://en.wikipedia.org/wiki/Cryptographic_nonce

package nonce

import (
	"fmt"
	"time"
	"crypto/sha1"
	"os"
	"io/ioutil"
	"strconv"
	"log"
)

var (
	// Used in every nonce hash generation. It is recommeneded that you
	// set this to a value unique to your application.
	Salt string = "go-vector-nonces"
	// The max age of a nonce before it expires. In nanoseconds.
	// Default value is 24 hours
	MaxAge int64 = 1e9 * 60 * 60 * 24;
	// The location where expired nonce data is stored
	SavePath = "./tmp/"
	// Prefix used to identify files created by this package.
	filePrefix string = "nonceegeto_"
	// How often the GC should run
	GcInterval int64 = 1e9 * 60 * 60//1.0 hrs
)

type Nonce struct {
	Nonce string
	Timestamp int64
}

// Generates a new nonce token based on the arguments & the package
// level salt. Arguments are handled using fmt.Sprintf.
func NewNonce(args ...interface{}) (n *Nonce) {
	return makeNonce(int64(time.Now().Nanosecond()), args)
}

// Recreates a nonce given a timestamp. If the recreated nonce does not match
// the one passed in by the user, it is invalid.
func RecreateNonce(timestamp int64, args ...interface{}) (n *Nonce) {
	return makeNonce(timestamp, args)
}

func makeNonce(timestamp int64, args ...interface{}) (n *Nonce) {
	str := fmt.Sprint(args...)
	str = fmt.Sprintf("a%s7xf6g%s%s9", Salt, str, timestamp)
	fmt.Println(str)
	hash := sha1.New()
	hash.Write([]byte(str))
	fmt.Println(hash.Sum(nil))
	nonce := fmt.Sprintf("%x", hash.Sum(nil))
	return &Nonce{Nonce:nonce, Timestamp:timestamp}
}

// Checks if a nonce is valid & can be used.
func IsValid(nonce *Nonce) bool {
	//If the time stamp is too old it is expired
	diff := int64(time.Now().Nanosecond()) - nonce.Timestamp
	if diff > MaxAge {
		return false
	}
	
	//if a file exists it has already been used
	//we expect an error to occur indicating the file doen't exist
	_, err := os.Stat(SavePath + filePrefix + nonce.Nonce);
	if(err == nil) {
		return false;
	}
	
	//otherwise it is ok
	return true
}

// Expires a nonce value so it cannot be used. Once expired calls to IsValid
// will return false.
func Expire(nonce *Nonce) (e *error) {
	file, err := os.Create(SavePath + filePrefix + nonce.Nonce)
	if err != nil {
		fmt.Println(err)
		return &err
	}
	
	// The comtents of ths file will be its time stamp. GC will use this value.
	timestamp := fmt.Sprintf("%d", nonce.Timestamp)
	_, err = file.WriteString(timestamp);
	if err != nil {
		fmt.Println(err)
		return &err
	}
	err = file.Close()
	
	return nil
}

// When a nonce is expired files are created to mark them so. `gc` runs
// to clean up these expired files after some time passes.
func gc() {
	for {
		//grab all the files in the save dir
		files, err := ioutil.ReadDir(SavePath)
		prefixLen := len(filePrefix)
		if err == nil {
			for _, fileInfo := range(files) {
				//look for files that start with our prefix
				name := fileInfo.Name();
				if len(name) > prefixLen && name[:prefixLen] == filePrefix {
					//the file content is expected to be a timestamp 
					content, err := ioutil.ReadFile(SavePath + name)
					if err != nil {
						log.Println(err)
						continue
					}
					timestamp, err := strconv.ParseInt(string(content), 10, 64)
					if(err == nil) {
						diff := int64(time.Now().Nanosecond()) - timestamp
						//if the file is old enough delete it
						if diff > MaxAge {
							err = os.Remove(SavePath + name)
							if err != nil {
								log.Println(err)
							}
						}
					} else {
						log.Println(err)
					}
				}
			}
		} else {
			log.Println(err)
		}
		//delay until the next interval
		time.Sleep(time.Duration(GcInterval))
	}
}

// init starts a goroutine to clean up old files every once in a while.
func init() {
	//go gc()
}

