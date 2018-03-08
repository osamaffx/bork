package main

import (
	"flag"
	"fmt"

	"github.com/westphae/bork/bork"
	"github.com/westphae/bork/config"
)

func main() {
	var configFile = flag.String("config", "./config.json", "config.json file to use")
	flag.Parse()
	err := config.ReadConfig(*configFile)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bork.Start()

	<-make(chan struct{})
	return
}
