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
	maxMembers = 30 // Maximum number of members in a fellowship
	arenaEnergyRate = 144 // Arena energy refreshes one point per 2.4 minutes = 144 seconds
	energyRate = 4*60 // Energy refreshes one point per 4 minutes = 240 seconds
	palantirRate = 4*60*60 // Palantir refreshes once per 4 hours
)

type scheduleItem struct {
	expireAt time.Time
	channel  chan struct{}
}

var (
	BotID        string
	borkSchedule map[string]scheduleItem
)

func Start() {
	borkSchedule = make(map[string]scheduleItem)

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

	fmt.Println("Bork is running!")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		c           chan struct{}
		d           time.Duration
		e           int
		err         error
		ackMessage  string
		doneMessage string
		helpMessage string
	)

	fmt.Sprintf(">> %s\n", m.ContentWithMentionsReplaced())

	if m.Author.ID == BotID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	if strings.HasPrefix("arena", f[0]) {
		helpMessage = fmt.Sprintf("%s, next time I'm gonna slice you up and feed you to the orclings " +
					"if you can't figure this out!\nJust tell me your current energy and I'll do the rest, " +
					"like this: %sarena 182.  It's simple, you dork!", m.Author.Mention(), config.BotPrefix)

		if len(f) != 2 {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		// If there's already a reminder, cancel it first.
		if i, ok := borkSchedule[m.Author.ID]; ok {
			close(i.channel)
			delete(borkSchedule, m.Author.ID)
		}

		e, err = strconv.Atoi(f[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		e = (300-e) * arenaEnergyRate - 60
		hour := int(e/3600)
		minute := int((e - 3600*hour)/60)
		second := int(e - 3600*hour - 60*minute)
		d = time.Duration(e) * time.Second
		ackMessage = fmt.Sprintf("Alright, %s, I'll rattle your cage in %d:%02d:%02d.  " +
				"Maybe I'll even let you out of it.", m.Author.Mention(), hour, minute, second)
		doneMessage = fmt.Sprintf(
			"Hey, you, %s!  It's about time to hit the arena, you lout.", m.Author.Mention())

		c = messageTimer(s, m, d, ackMessage, doneMessage)
		borkSchedule[m.Author.ID] = scheduleItem{
			time.Now().Add(d),
			c}
		fmt.Println(borkSchedule)
	}
}

//TODO: timers for Palantir, campaign energy, ability refresh, hourly warlord, daily warlord
//TODO: keep a struct of user data
//TODO: say the local time too (or UTC if user doesn't give tz info)
//TODO: vary the messages sent
//TODO: orc adviser
//TODO: inscription adviser (who needs which)
func messageTimer(s *discordgo.Session, m *discordgo.MessageCreate, e time.Duration,
	ackMessage string, doneMessage string) (c chan struct{}){

	c = make(chan struct{}, 1)

	s.ChannelMessageSend(m.ChannelID, ackMessage)

	go func() {
		t := time.After(e)
		select {
		case <-t:
			s.ChannelMessageSend(m.ChannelID, doneMessage)
		case <-c:
			fmt.Sprintf("! Canceling %s reminder", m.Author)
		}
	}()

	return c
}
