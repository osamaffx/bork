package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
)

// userInfo contains information on the players
// values in a map[string]userInfo where the key is the Discord ID
type userInfo struct {
	TimeZone   string `json:"tz"`          // Time zone for user for reporting times
	MaxEnergy  int    `json:"max_energy"`  // Max energy of user (i.e. 174 for lvl 80)
	MaxAbility int    `json:"max_ability"` // Max ability points of user (default 12)
	Uses       int    `json:"Uses"`        // Number of times user has called Bork
}

var users map[string]userInfo

func usersSetup() {
	users = make(map[string]userInfo)

	loadUsers("./data/users.json")
}

func profileHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var (
		profile     userInfo
		ok          bool
		err         error
		helpMessage string
		i           int
		v           int
	)

	if m.Author.ID == BorkID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}
	fmt.Printf(">> %s\n", m.ContentWithMentionsReplaced())

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"GMT", 144, 12, 0}
	}
	profile.Uses += 1

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	if !strings.HasPrefix("profile", f[0]) {
		return
	}

	helpMessage = fmt.Sprintf("%s, you scumbag, cough up your info: time zone, max campaign energy, " +
		"and number of ability point refreshes, like this: %sprofile EDT energy 178 ability 14.  If you don't " +
		"tell me one of them, I'll use some default values.  They can be in any order.\n" +
		"Your current profile is: time zone %s, campaign energy %d, ability points %d.",
		m.Author.Mention(), config.BotPrefix, profile.TimeZone, profile.MaxEnergy, profile.MaxAbility)

	if len(f) > 6 {
		s.ChannelMessageSend(m.ChannelID, helpMessage)
		return
	}

	i = 1
	for i < len(f){
		if strings.HasPrefix("energy", f[i]) {
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
		} else if strings.HasPrefix("ability", f[i]) {
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
		} else {
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

	users[m.Author.ID] = profile
	saveUsers("./data/users.json")
}
