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
	energyRate = 4*60 // Energy refreshes one point per 4 minutes = 240 seconds
	arenaEnergyRate = 144 // Arena energy refreshes one point per 2.4 minutes = 144 seconds
	palantirRate = 4*60*60 // Palantir refreshes once per 4 hours
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

	goBot.AddHandler(messageHandler)

	err = goBot.Open()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running!")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		e           int
		err         error
		ackMessage  string
		doneMessage string
	)

	fmt.Println(m.ContentWithMentionsReplaced())

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

		e = int((299.5-float32(e)) * arenaEnergyRate)
		hour := int(e/3600)
		minute := int((e - 3600*hour)/60)
		second := int(e - 3600*hour - 60*minute)
		ackMessage = fmt.Sprintf("Alright, %s, I'll rattle your cage in %d:%02d:%02d.  " +
				"Maybe I'll even let you out of it.", m.Author.Mention(), hour, minute, second)
		doneMessage = fmt.Sprintf(
			"Hey, you, %s, WAKE UP!  It's time to hit the arena, you lout.", m.Author.Mention())

		go messageTimer(s, m, e, ackMessage, doneMessage)

	}
}

//TODO: return a channel for resetting time
//TODO: restrict to specific bot channel
//TODO: timers for palantir, campaign energy, ability refresh, hourly warlord, daily warlord
//TODO: keep a struct of user data
//TODO: say the local time too (or UTC if user doesn't give tz info)
//TODO: vary the messages sent
//TODO: orc adviser
//TODO: inscription adviser (who needs which)
func messageTimer(s *discordgo.Session, m *discordgo.MessageCreate, e int, ackMessage string, doneMessage string) {
	s.ChannelMessageSend(m.ChannelID, ackMessage)
	time.Sleep(time.Duration(e) * time.Second)
	s.ChannelMessageSend(m.ChannelID, doneMessage)
}
