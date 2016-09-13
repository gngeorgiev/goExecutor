package main

import (
	"flag"

	"log"

	"github.com/gngeorgiev/goExecutor/apiServer"
)

var (
	image   string
	file    string
	command string
)

func init() {
	flag.StringVar(&image, "image", "", "Docker image to run into")
	flag.StringVar(&file, "file", "", "The file which contains the code to run")
	flag.StringVar(&command, "command", "", "The command to run when executing the code")
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	//flag.Parse()

	//if image == "" {
	//	log.Fatal("Specify a docker image to run the code into")
	//}
	//
	//if file == "" {
	//	log.Fatal("Specify a file with code to run")
	//}
	//
	//if command == "" {
	//	log.Fatal("Specify a command to execute the code")
	//}

	log.Fatal(apiServer.Listen())
}
