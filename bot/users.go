package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/westphae/bork/config"
)

// userInfo contains information on the players
// values in a map[string]userInfo where the key is the Discord ID
type userInfo struct {
	TimeZone      string `json:"tz"`             // Time zone for user for reporting times
	MaxEnergy     int    `json:"max_energy"`     // Max energy of user (i.e. 174 for lvl 80)
	MaxAbility    int    `json:"max_ability"`    // Max ability points of user (default 12)
	MaxDominance  int    `json:"max_dominance"`  // Hourly rate at which dominance builds
	DominanceRate int    `json:"dominance_rate"` // Hourly rate at which dominance builds
	Uses          int    `json:"Uses"`           // Number of times user has called Bork
}

var users map[string]userInfo

func usersSetup(s *discordgo.Session) {
	users = make(map[string]userInfo)

	loadUsers("./data/users.json")
}

func profileHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var (
		profile     userInfo
		ok          bool
		err         error
		i           int
		v           int
	)

	if m.Author.ID == BorkID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"-5", maxEnergy, maxAbility, maxDominance, dominanceRate, 0}
	}
	profile.Uses += 1

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	if !strings.HasPrefix("profile", f[0]) {
		return
	}

	if len(f) > 6 {
		sendHelpMessage(s)
		return
	}

	i = 1
	for i < len(f){
		if strings.HasPrefix("energy", f[i]) {
			if len(f) <= i+1{
				sendHelpMessage(s)
				return
			}
			v, err = strconv.Atoi(f[i+1])
			if err != nil{
				sendHelpMessage(s)
				return
			}
			profile.MaxEnergy = v
			i += 2
		} else if strings.HasPrefix("ability", f[i]) {
			if len(f) <= i+1{
				sendHelpMessage(s)
				return
			}
			v, err = strconv.Atoi(f[i+1])
			if err != nil{
				sendHelpMessage(s)
				return
			}
			profile.MaxAbility = v
			i += 2
		} else if strings.HasPrefix("dominance", f[i]) {
			if len(f) <= i+1{
				sendHelpMessage(s)
				return
			}
			v, err = strconv.Atoi(f[i+1])
			if err != nil{
				sendHelpMessage(s)
				return
			}
			profile.MaxDominance = v
			i += 2
		} else if strings.HasPrefix("rate", f[i]) {
			if len(f) <= i+1{
				sendHelpMessage(s)
				return
			}
			v, err = strconv.Atoi(f[i+1])
			if err != nil{
				sendHelpMessage(s)
				return
			}
			profile.DominanceRate = v
			i += 2
		} else {
			/*
			_, err := time.LoadLocation(f[i])
			if err != nil{
				sendHelpMessage(s)
				return
			}
			*/
			profile.TimeZone = f[i]
			i += 1
		}
	}

	s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("Here's your new info, %s: time zone is %s, max energy is %d, max ability points is %d, max dominance is %d, dominance rate is %d\n",
			m.Author.Mention(), profile.TimeZone, profile.MaxEnergy, profile.MaxAbility, profile.MaxDominance, profile.DominanceRate))

	users[m.Author.ID] = profile
	saveUsers("./data/users.json")
}
