package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
)

type scheduleItem struct {
	expireAt time.Time
	channel  chan struct{}
}

const (
	arenaEnergyRate = 144 // Arena energy refreshes one point per 2.4 minutes = 144 seconds
	energyRate = 4*60 // Energy refreshes one point per 4 minutes = 240 seconds
	abilityRate = 5*60 // Ability points refresh one point per 5 minutes = 300 seconds
	palantirRate = 4*60*60 // Palantir refreshes once per 4 hours
)

var (
	arenaSchedule   map[string]scheduleItem
	energySchedule  map[string]scheduleItem
	abilitySchedule map[string]scheduleItem
)

func timersSetup() {
	arenaSchedule = make(map[string]scheduleItem)
	energySchedule = make(map[string]scheduleItem)
	abilitySchedule = make(map[string]scheduleItem)

	//loadSchedule("./data/arena-schedule.json")
}

func timerHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		profile     userInfo
		c           chan struct{}
		d           time.Duration
		e           int
		ok          bool
		err         error
		ackMessage  string
		doneMessage string
		helpMessage string
	)

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"GMT", 144, 12, 0}
	}
	profile.Uses += 1

	if m.Author.ID == BorkID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}
	fmt.Printf(">> %s\n", m.ContentWithMentionsReplaced())

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	switch {
	case strings.HasPrefix("arena", f[0]):
		helpMessage = fmt.Sprintf("%s, next time I'm gonna slice you up and feed you to the orclings " +
			"if you can't figure this out!\nJust tell me your current energy and I'll do the rest, " +
			"like this: %sarena 182.  It's simple, you dork!", m.Author.Mention(), config.BotPrefix)

		if len(f) > 2 {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		// If there's already a reminder, cancel it first.
		if i, ok := arenaSchedule[m.Author.ID]; ok{
			close(i.channel)
			delete(arenaSchedule, m.Author.ID)
		}

		e, err = strconv.Atoi(f[1])
		if err != nil{
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
		arenaSchedule[m.Author.ID] = scheduleItem{
			time.Now().Add(d),
			c}

		//saveSchedule("./data/arena_schedule.json")

	case strings.HasPrefix("energy", f[0]):
		helpMessage = fmt.Sprintf("%s, next time I'm gonna slice you up and feed you to the orclings " +
			"if you can't figure this out!\nJust tell me your current energy and I'll do the rest.  " +
			"You can also tell me your max if I don't already know it.  Just do it like this: " +
			"%senergy 102 174.  It's simple, you dork!", m.Author.Mention(), config.BotPrefix)

		if len(f) < 2 || len(f) > 3 {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		// If there's already a reminder, cancel it first.
		if i, ok := energySchedule[m.Author.ID]; ok{
			close(i.channel)
			delete(energySchedule, m.Author.ID)
		}

		e, err = strconv.Atoi(f[1])
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		if len(f) == 3 {
			max, err := strconv.Atoi(f[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, helpMessage)
				return
			}
			profile.MaxEnergy = max
		}

		e = (profile.MaxEnergy-e) * energyRate - 60
		hour := int(e/3600)
		minute := int((e - 3600*hour)/60)
		second := int(e - 3600*hour - 60*minute)
		d = time.Duration(e) * time.Second
		ackMessage = fmt.Sprintf("Alright, %s, I'll rattle your cage in %d:%02d:%02d.  " +
			"Maybe I'll even let you out of it.", m.Author.Mention(), hour, minute, second)
		doneMessage = fmt.Sprintf(
			"Hey, you, %s!  Your campaign energy is just about full, you lazy bum!", m.Author.Mention())

		c = messageTimer(s, m, d, ackMessage, doneMessage)
		energySchedule[m.Author.ID] = scheduleItem{
			time.Now().Add(d),
			c}

		//saveSchedule("./data/energy_schedule.json")

	case strings.HasPrefix("ability", f[0]):
		helpMessage = fmt.Sprintf("%s, next time I'm gonna slice you up and feed you to the orclings " +
			"if you can't figure this out!\nJust tell me your current ability points and I'll do the rest.  " +
			"You can also tell me your max if I don't already know it.  Just do it like this: " +
			"%sability 2 13.  It's simple, you dork!", m.Author.Mention(), config.BotPrefix)

		if len(f) < 2 || len(f) > 3 {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		// If there's already a reminder, cancel it first.
		if i, ok := abilitySchedule[m.Author.ID]; ok{
			close(i.channel)
			delete(abilitySchedule, m.Author.ID)
		}

		e, err = strconv.Atoi(f[1])
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		if len(f) == 3 {
			max, err := strconv.Atoi(f[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, helpMessage)
				return
			}
			profile.MaxAbility = max
		}

		e = (profile.MaxAbility-e) * abilityRate - 60
		hour := int(e/3600)
		minute := int((e - 3600*hour)/60)
		second := int(e - 3600*hour - 60*minute)
		d = time.Duration(e) * time.Second
		ackMessage = fmt.Sprintf("Alright, %s, I'll rattle your cage in %d:%02d:%02d.  " +
			"Maybe I'll even let you out of it.", m.Author.Mention(), hour, minute, second)
		doneMessage = fmt.Sprintf(
			"Hey, you, %s!  Your ability points are just about full, you lazy bum!", m.Author.Mention())

		c = messageTimer(s, m, d, ackMessage, doneMessage)
		abilitySchedule[m.Author.ID] = scheduleItem{
			time.Now().Add(d),
			c}

		//saveSchedule("./data/ability_schedule.json")
	}

	users[m.Author.ID] = profile
	saveUsers("./data/users.json")
}

//TODO: display times until all ppl are going
//TODO: timers for Palantir
//TODO: timer for orc jobs / misc
//TODO: timers for warlords
//TODO: timers for requests
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
