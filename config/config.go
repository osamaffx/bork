package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	// Public variables
	Token      string
	BotPrefix  string
	BotChannel string

	// Private variables
	config *configStruct
)

type configStruct struct {
	Token      string `json:"Token"`
	BotPrefix  string `json:"BotPrefix"`
	BotChannel string `json:"BotChannel"`
}

func ReadConfig(fn string) error {
	fmt.Sprintf("Reading config file %s\n", fn)

	file, err := ioutil.ReadFile(fn)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	BotChannel = config.BotChannel

	return nil
}
