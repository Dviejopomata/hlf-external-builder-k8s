package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	log.Printf("Running")
	msg := fmt.Sprintf("This was executed at %s", time.Now().Format(time.RFC3339))
	err := ioutil.WriteFile("/tmp/run.txt", []byte( msg), 777)
	if err != nil {
		log.Printf("Failed to write file")
		os.Exit(1)
	}
	os.Exit(0)
}
