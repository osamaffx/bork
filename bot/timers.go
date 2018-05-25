package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
	"log"
)

type scheduleItem struct {
	expireAt time.Time
	channel  chan struct{}
}

const (
	arenaEnergyRate = 144 // Arena energy refreshes one point per 2.4 minutes = 144 seconds
	maxArenaEnergy  = 300
	energyRate      = 4 * 60 // Energy refreshes one point per 4 minutes = 240 seconds
	maxEnergy       = 178
	abilityRate     = 5 * 60 // Ability points refresh one point per 5 minutes = 300 seconds
	timeRate        = 60 // Generic convention is time in minutes = 60 seconds
	maxAbility      = 12
)

var (
	schedule map[string]map[string]scheduleItem
)

func timersSetup(s *discordgo.Session) {
	schedule = make(map[string]map[string]scheduleItem)
	data, err := loadSchedule("./data/schedule.json")
	if err != nil {
		log.Println(err.Error())
		return
	}

	for id, v := range data {
		user, err := s.User(id)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		schedule[id] = make(map[string]scheduleItem)

		for sType, t := range v {
			d := t.Sub(time.Now())
			if d < 0 {
				log.Printf("Too late to run %s timer at %s\n", sType, t)
				continue
			}

			c := messageTimer(s, config.BotChannel, d,
				fmt.Sprintf("%s, you have full %s.\n", user.Mention(), sType),
			)

			schedule[id][sType] = scheduleItem{time.Now().Add(d), c}
		}
	}
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
		sType, msg   string
		rate         int
		max, newMax  int
		offset       int
	)

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"GMT", maxEnergy, maxAbility, 0}
	}
	profile.Uses += 1

	if m.Author.ID == BorkID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}

	f := strings.SplitN(m.Content[1:len(m.Content)], " ", 2)

	if userSchedule, ok = schedule[m.Author.ID]; !ok {
		schedule[m.Author.ID] = make(map[string]scheduleItem)
	}

	switch {
	case strings.HasPrefix("arena", f[0]) && len(f)>1:
		sType = "arena energy"
		msg = fmt.Sprintf("you have full %s", sType)
		rate = arenaEnergyRate
		offset = 60
		if e, newMax, err = parseCurMax(f[1]); err != nil {
			log.Print(err)
			sendHelpMessage(s)
			return
		}
		if newMax > 0 {
			max = newMax
		} else {
			max = maxArenaEnergy
		}
	case strings.HasPrefix("energy", f[0]) && len(f)>1:
		sType = "campaign energy"
		msg = fmt.Sprintf("you have full %s", sType)
		rate = energyRate
		offset = 60
		if e, newMax, err = parseCurMax(f[1]); err != nil {
			log.Print(err)
			sendHelpMessage(s)
			return
		}
		if newMax > 0 {
			profile.MaxEnergy = newMax
		}
		max = profile.MaxEnergy
	case strings.HasPrefix("ability", f[0]) && len(f)>1:
		sType = "ability points"
		msg = fmt.Sprintf("you have full %s", sType)
		rate = abilityRate
		offset = 60
		if e, newMax, err = parseCurMax(f[1]); err != nil {
			log.Print(err)
			sendHelpMessage(s)
			return
		}
		if newMax > 0 {
			profile.MaxAbility = newMax
		}
		max = profile.MaxAbility
	case strings.HasPrefix("time", f[0]) && len(f)>1:
		rate = timeRate
		offset = 0
		if max, sType, err = parseTime(f[1]); err != nil {
			log.Print(err)
			sendHelpMessage(s)
			return
		}
		msg = fmt.Sprintf("your timer for %s has expired", sType)
		e = 0
	default:
		sendHelpMessage(s)
		return
	}

	// If there's already a reminder, cancel it first.
	if i, ok := userSchedule[sType]; ok {
		close(i.channel)
		delete(userSchedule, sType)
	}

	e = (max-e)*rate - offset
	hour := int(e / 3600)
	minute := int((e - 3600*hour) / 60)
	second := int(e - 3600*hour - 60*minute)
	d = time.Duration(e) * time.Second

	s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("%s, you will receive an alert for %s in %d:%02d:%02d.\n",
			m.Author.Mention(), sType, hour, minute, second),
	)

	c = messageTimer(s, m.ChannelID, d,
		fmt.Sprintf("%s, %s.\n", m.Author.Mention(), msg),
	)
	schedule[m.Author.ID][sType] = scheduleItem{time.Now().Add(d), c}

	users[m.Author.ID] = profile
	saveUsers("./data/users.json")
	saveSchedule("./data/schedule.json")
}

// parseCurMax returns an int of current value and max value given an input string.
func parseCurMax(input string) (cur int, max int, err error) {
	f := strings.Split(input, " ")

	cur, err = strconv.Atoi(f[0])
	if err != nil {
		return 0, 0, err
	}

	if len(f) == 2 {
		max, err = strconv.Atoi(f[1])
		if err != nil {
			return 0, 0, err
		}
	}

	return cur, max, nil
}

// parseTime returns an int of minutes and a message given a string like "h:mm mes sage".
func parseTime(input string) (output int, msg string, err error) {
	if len(input)==0 {
		err = fmt.Errorf("bad string passed: %s", input)
		return
	}

	f := strings.SplitN(input, " ", 2)

	if len(f)==2 {
		msg = f[1]
	} else {
		msg = "whatever the hell you wanted"
	}

	if output, err = strconv.Atoi(f[0]); err == nil {
		return
	}

	if i := strings.Index(f[0], ":"); i > 0 {
		h, err := strconv.Atoi(f[0][:i])
		if err != nil {
			return 0, "", err
		}
		m, err := strconv.Atoi(f[0][i+1:])
		if err != nil {
			return 0, "", err
		}
		output = 60*h + m
		return output, msg, nil
	}
	return 0, "", fmt.Errorf("error parsing time %s", f[0])
}

//TODO: arbitrary timer by h:mm (fortress refreshes, palantir, orc jobs, requests)
//TODO: timers for warlords
//TODO: display times until all ppl are going
//TODO: say the local time too (or UTC if user doesn't give tz info)
//TODO: vary the messages sent
//TODO: orc adviser
//TODO: inscription adviser (who needs which)
func messageTimer(s *discordgo.Session, channelID string, e time.Duration,
	doneMessage string) (c chan struct{}) {

	c = make(chan struct{}, 1)

	go func() {
		t := time.After(e)
		select {
		case <-t:
			s.ChannelMessageSend(channelID, doneMessage)
		case <-c:
		}
	}()

	return c
}
