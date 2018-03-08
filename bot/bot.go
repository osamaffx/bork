package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
	"time"
	"strconv"
	"encoding/json"
	"io/ioutil"
)

const (
	maxMembers = 30 // Maximum number of members in a fellowship
	arenaEnergyRate = 144 // Arena energy refreshes one point per 2.4 minutes = 144 seconds
	energyRate = 4*60 // Energy refreshes one point per 4 minutes = 240 seconds
	palantirRate = 4*60*60 // Palantir refreshes once per 4 hours
)

// userInfo contains information on the players
// values in a map[string]userInfo where the key is the Discord ID
type userInfo struct {
	TimeZone   string `json:"tz"`        // Time zone for user for reporting times
	MaxEnergy  int    `json:max_energy`  // Max energy of user (i.e. 174 for lvl 80)
	MaxAbility int    `json:max_ability` // Max ability points of user (default 12)
	Uses       int    `json:Uses`        // Number of times user has called Bork
}

type scheduleItem struct {
	expireAt time.Time
	channel  chan struct{}
}

var (
	BotID           string
	arenaSchedule   map[string]scheduleItem
	users           map[string]userInfo
)

func Start() {
	users = make(map[string]userInfo)
	arenaSchedule = make(map[string]scheduleItem)

	loadUsers("./data/users.json")
	//loadSchedule("./data/arena-schedule.json")

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

	fmt.Printf(">> %s\n", m.ContentWithMentionsReplaced())

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"GMT", 144, 12, 0}
	}
	profile.Uses += 1

	if m.Author.ID == BotID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}

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


	case strings.HasPrefix("profile", f[0]):
		var (
			i          int
			v          int
		)

		helpMessage = fmt.Sprintf("%s, you scumbag, cough up your info: time zone, max campaign energy, " +
			"and number of ability point refreshes, like this: %sprofile EDT energy 178 ability 14.  If you don't " +
			"tell me one of them, I'll use some default values.  They can be in any order.\n" +
			"Your current profile is: time zone %s, campaign energy %d, ability points %d.",
			m.Author.Mention(), config.BotPrefix, profile.TimeZone, profile.MaxEnergy, profile.MaxAbility)

		if len(f) > 6{
			fmt.Printf("You entered %d arguments", len(f))
			s.ChannelMessageSend(m.ChannelID, helpMessage)
			return
		}

		i = 1
		for i < len(f){
			if strings.HasPrefix("energy", f[i]){
				if len(f) <= i+1{
					s.ChannelMessageSend(m.ChannelID, helpMessage)
					return
				}
				v, err = strconv.Atoi(f[i+1])
				if err != nil{
					s.ChannelMessageSend(m.ChannelID, helpMessage)
					return
				}
				profile.MaxEnergy = v
				i += 2
			} else if strings.HasPrefix("ability", f[i]){
				if len(f) <= i+1{
					s.ChannelMessageSend(m.ChannelID, helpMessage)
					return
				}
				v, err = strconv.Atoi(f[i+1])
				if err != nil{
					s.ChannelMessageSend(m.ChannelID, helpMessage)
					return
				}
				profile.MaxAbility = v
				i += 2
			} else{
				_, err := time.LoadLocation(f[i])
				if err != nil{
					s.ChannelMessageSend(m.ChannelID, helpMessage)
					return
				}
				profile.TimeZone = f[i]
				i += 1
			}
		}

		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Here's your new info, %s: time zone is %s, max energy is %d, max ability points is %d\n",
				m.Author.Mention(), profile.TimeZone, profile.MaxEnergy, profile.MaxAbility))
	}

	users[m.Author.ID] = profile
	saveUsers("./data/users.json")
}

// loadUsers retrieves the Users struct from a file.
func loadUsers(filename string) (err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading %s file: %s\n", filename, err.Error())
	} else {
		err = json.Unmarshal(b, &users)
		if err != nil {
			fmt.Printf("Error unmarshaling %s: %s\n", filename, err.Error())
		}
	}
	return
}

// saveUsers saves the Users struct to a file.
func saveUsers(filename string) (err error) {
	b, err := json.Marshal(users)
	if err != nil {
		fmt.Printf("Error marshaling %s: %s\n", filename, err.Error())
		return
	}
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		fmt.Printf("Error writing %s file: %s\n", filename, err.Error())
		return
	}
	return
}

func loadSchedule() (err error) {
	// type (arena, warlord, ability, energy, palantir)
	// Key: discordgo.User.ID
	// Value: expiry_time
	var (
		b []byte
		data map[string]time.Time
	)

	if b, err = ioutil.ReadFile("./borkSchedule.json"); err != nil {
		fmt.Printf("Error reading borkSchedule.json file: %s\n", err.Error())
		return
	}
	if err = json.Unmarshal(b, &data); err != nil {
		fmt.Printf("Error unmarshaling borkSchedule.json: %s\n", err.Error())
		return
	}

	return
}

func saveSchedule() (err error) {
	// Write current data to file in case we have to restart
	b, err := json.Marshal(borkSchedule)
	if err != nil {
		fmt.Printf("Error marshaling borkSchedule.json: %s\n", err.Error())
		return
	}
	err = ioutil.WriteFile("./borkSchedule.json", b, 0644)
	if err != nil {
		fmt.Printf("Error writing borkSchedule.json file: %s\n", err.Error())
		return
	}
	return
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
