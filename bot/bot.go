package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
	"time"
	"strconv"
)

const (
	energyRate = 240 // Energy refreshes one point per 4 minutes = 240 seconds
	arenaEnergyRate = 144 // Arena energy refreshes one point per 2.4 minutes = 144 seconds
)

var BotID string

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

	BotID = u.ID

	goBot.AddHandler(arenaMessageHandler)
	goBot.AddHandler(energyMessageHandler)

	err = goBot.Open()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running!")
}

func arenaMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		e    int
		err  error
	)
	if m.Author.ID == BotID || !strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	if f[0] == "arena" {
		bye := func() {
			s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("%s, next time I'm gonna slice you up and feed you to the orclings " +
					"if you can't figure this out!\nJust tell me your current energy and I'll do the rest, " +
					"like this: %sarena 182.  It's simple, you dork!",
					m.Author.Mention(), config.BotPrefix))
		}

		if len(f) != 2 {
			bye()
			return
		}

		e, err = strconv.Atoi(f[1])
		if err != nil {
			bye()
			return
		}

		go func() {
			e = (300-e) * arenaEnergyRate
			hour := int(e/3600)
			minute := int((e - 3600*hour)/60)
			second := int(e - 3600*hour - 60*minute)
			s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("Alright, @%s, I'll rattle your cage in %d:%02d:%02d.  " +
					"Maybe I'll even let you out of it.", m.Author.Mention(), hour, minute, second))
			time.Sleep(time.Duration(e) * time.Second)
			_, _ = s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("Hey, you, %s, WAKE UP!  It's time to hit the arena, you lout.", m.Author.Mention()))
		}()

	}
}

func energyMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		e, max int
		err    error
	)
	if m.Author.ID == BotID {
		return
	}

	if strings.HasPrefix(m.Content, config.BotPrefix) {

		f := strings.Split(m.Content[1:len(m.Content)], " ")

		if f[0] == "energy" {
			e, err = strconv.Atoi(f[1])
			if err != nil {
				fmt.Println(err.Error())
				s.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("%s, next time I'm gonna slice you up and eat you myself if you keep talking trash to me!",
						m.Author.Mention()))
				return
			}

			if len(f) > 1 {
				max, err = strconv.Atoi(f[2])
				if err != nil {
					fmt.Println(err.Error())
					s.ChannelMessageSend(m.ChannelID,
						fmt.Sprintf("%s, next time I'm gonna slice you up and feed you to my caragors if you don't give me a valid maximum energy!",
							m.Author.Mention()))
					return
				}
			} else {
				max = 178 // Max energy for level 80
			}

			go func() {
				e = (max-e) * arenaEnergyRate
				s.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("Alright, @%s, I'll rattle your cage in about %d minutes.", m.Author.Mention(), int(e/60)))
				time.Sleep(time.Duration(e) * time.Second)
				_, _ = s.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("Hey, you, %s, WAKE UP!  It's time to hit the arena, you lazy bastard.", m.Author.Mention()))
			}()
		}
	}
}
