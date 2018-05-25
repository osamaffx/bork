package bot

import (
	"fmt"
	"github.com/westphae/bork/config"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func sendHelpMessage(s *discordgo.Session) (err error) {
	var helpMessage string

	helpMessage = fmt.Sprintf("```asciidoc" + `
Bork is a bot who can help you remember to do things at the right time in Shadow of War Mobile.

The currently supported commands are "profile", "arena", "energy", "ability" and "time".  When you enter a command, you only have to enter enough of the letters to uniquely specify it.

Your profile has information about you including your offset from GMT for telling you times (not yet implemented), your max campaign energy (determined by your level), and your max ability points (which can go up with VIP level).  Once you specify one of these values either with the profile command or another command, it will be remembered.
	%[1]sprofile [<GMT offset=-5>] [energy <max campaign energy=%[3]d>] [ability <max ability points=%[4]d>]

The following timer commands will give you a roughly one-minute notice.  Entering the max value is optional and will be remembered (except arena, which will always be 300 unless you specify a lower value for just the current run).
	%[1]sarena <current arena energy=0> [<max arena energy=%[2]d>]
	%[1]senergy <current campaign energy=0> [<max campaign energy=%[3]d>]
	%[1]sability <current ability points=0> [<max ability points=%[4]d>]

The "time" command is  customizable and will send you whatever message you specify at whatever time you want it to be delivered. Time can either be specified as an integer number of minutes or as h:mm.  Suggested uses: palantir, fortress refreshes, request refreshes, orc jobs.
    %[1]stime <time> [message]
` + "```",
		config.BotPrefix, maxArenaEnergy, maxEnergy, maxAbility)

	if _, err = s.ChannelMessageSend(config.BotChannel, helpMessage); err != nil {
		return
	}

	helpMessage = fmt.Sprintf("```asciidoc" + `
Examples:
	%[1]sarena 12 150    - will tell you when your arena energy reaches 150.  
	%[1]sarena 18        - will tell you when your arena energy reaches 300, even if you previously had an alarm set for 150.
	%[1]se 88 174        - will tell you when your campaign energy reaches 174 from 88 and also permanently set your max campaign energy to 174.
	%[1]sab 4 13         - will tell you when your ability points refresh from 4 up to 13, and also permanently set your max ability ponits to 13.
	%[1]sprofile         - will tell you your current profile.
    %[1]spr e 176 a 14   - will set your profile to max campaign energy of 176 and max ability points of 14.
    %[1]st 5 skirmish    - will tell you when your skirmish orc job is finished after 5 minutes.
    %[1]st 7:55 request  - will tell you when you can request more items 
    %[1]st 11:55 MM Norm - will tell you when Minas Morgul Normal tower is about to refresh
` + "```",
		config.BotPrefix, maxArenaEnergy, maxEnergy, maxAbility)

	_, err = s.ChannelMessageSend(config.BotChannel, helpMessage)
	return
}

func helpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var (
		profile     userInfo
		ok          bool
	)

	if m.Author.ID == BorkID || (config.BotChannel != "" && m.ChannelID != config.BotChannel) ||
		!strings.HasPrefix(m.Content, config.BotPrefix) {
		return
	}

	if profile, ok = users[m.Author.ID]; !ok {
		profile = userInfo{"-5", 144, 12, 0}
	}
	profile.Uses += 1

	f := strings.Split(m.Content[1:len(m.Content)], " ")

	if !strings.HasPrefix("help", f[0]) {
		return
	}

	sendHelpMessage(s)
	return
}
