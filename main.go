package main

import (
	"fmt"

	"github.com/westphae/bork/bot"
	"github.com/westphae/bork/config"
)

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bot.Start()

	<-make(chan struct{})
	return
}
