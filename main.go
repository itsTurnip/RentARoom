package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var token string
var appID string

func main() {
	flag.StringVar(&token, "token", "", "Bot token")
	flag.StringVar(&appID, "appid", "", "Application ID")
	flag.Parse()
	if token == "" || appID == "" {
		fmt.Println("App ID and Token couldn't be empty")
		return
	}
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	bot.AddHandler(messageCreate)

	bot.AddHandler(func(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
		channel, ok := rented[vsu.UserID]
		if ok {
			if vsu.ChannelID != channel.ID {
				_, err := s.ChannelDelete(channel.ID)
				if err != nil {
					fmt.Println(err)
				}
				delete(rented, vsu.UserID)
			} else {
				channel.visited = true
			}
		}
	})

	err = bot.Open()
	if err != nil {
		panic(err)
	}
	fmt.Println("Bot running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	for _, v := range rented {
		bot.ChannelDelete(v.ID)
	}
	bot.Close()
}
