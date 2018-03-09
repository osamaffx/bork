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
	maxArenaEnergy = 300
	energyRate = 4*60 // Energy refreshes one point per 4 minutes = 240 seconds
	maxEnergy = 178
	abilityRate = 5*60 // Ability points refresh one point per 5 minutes = 300 seconds
	maxAbility = 12
	palantirRate = 4*60*60 // Palantir refreshes once per 4 hours
	maxPalantir = 1
)

var (
	schedule map[string]map[string]scheduleItem
)

func timersSetup() {
	schedule = make(map[string]map[string]scheduleItem)

	//loadSchedule("./data/schedule.json")
}

func timerHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		profile      userInfo
		userSchedule map[string]scheduleItem
		c            chan struct{}
		d            time.Duration
		e            int
		ok           bool
		err          error
		sType        string
		rate         int
		max, newMax  int
		helpMessage  string
	)

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"GMT", maxEnergy, maxAbility, 0}
	}
	profile.Uses += 1

	helpMessage = fmt.Sprintf("You can enter these commands:\n" +
		"%[1]sa[rena] <current arena energy> [<max arena energy=300>]\n" +
		"%[1]se[nergy] <current campaign energy> [<max campaign energy=178>]\n" +
		"%[1]sab[ility] <current ability points> [<max ability points=12>]\n",
		config.BotPrefix)

	if m.Author.ID == BorkID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}
	fmt.Printf(">> %s\n", m.ContentWithMentionsReplaced())

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	if userSchedule, ok = schedule[m.Author.ID]; !ok {
		schedule[m.Author.ID] = make(map[string]scheduleItem)
	}

	if len(f) < 2 || len(f) > 3 {
		s.ChannelMessageSend(m.ChannelID, helpMessage )
		return
	}

	e, err = strconv.Atoi(f[1])
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, helpMessage)
		return
	}

	if len(f) == 3 {
		newMax, err = strconv.Atoi(f[2])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}
	}

	switch {
	case strings.HasPrefix("arena", f[0]):
		sType = "arena energy"
		rate = arenaEnergyRate
		if newMax > 0 {
			max = newMax
		} else {
			max = maxArenaEnergy
		}
	case strings.HasPrefix("energy", f[0]):
		sType = "campaign energy"
		rate = energyRate
		if newMax > 0 {
			profile.MaxEnergy = newMax
		}
		max = profile.MaxEnergy
	case strings.HasPrefix("ability", f[0]):
		sType = "ability points"
		rate = abilityRate
		if newMax > 0 {
			profile.MaxAbility = newMax
		}
		max = profile.MaxAbility
	}

	// If there's already a reminder, cancel it first.
	if i, ok := userSchedule[sType]; ok {
		close(i.channel)
		delete(userSchedule, sType)
	}

	e = (max-e) * rate - 60
	hour := int(e/3600)
	minute := int((e - 3600*hour)/60)
	second := int(e - 3600*hour - 60*minute)
	d = time.Duration(e) * time.Second

	c = messageTimer(s, m, d,
		fmt.Sprintf("%s, you will receive an alert for %s in %d:%02d:%02d.",
			m.Author.Mention(), sType, hour, minute, second),
		fmt.Sprintf("%s, you have full %s.", m.Author.Mention(), sType),
	)
	schedule[m.Author.ID][sType] = scheduleItem{time.Now().Add(d),c}

	users[m.Author.ID] = profile
	saveUsers("./data/users.json")
	//saveSchedule("./data/schedule.json")
}

//TODO: timers for Palantir
//TODO: timer for orc jobs / misc
//TODO: timers for warlords
//TODO: timers for requests
//TODO: display times until all ppl are going
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
