package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
)

var BorkID string

func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
	}

	BorkID = u.ID

	usersSetup()
	goBot.AddHandler(profileHandler)

	timersSetup()
	goBot.AddHandler(timerHandler)

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bork is running!")
}
