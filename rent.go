package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type rentChannel struct {
	guild   string
	owner   string
	ID      string
	visited bool
}

var rented = make(map[string]*rentChannel)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID &&
		len(m.Mentions) != 0 &&
		m.Mentions[0].ID != s.State.User.ID {
		return
	}

	channel, err := s.Channel(m.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if channel.Type != discordgo.ChannelTypeGuildText {
		return
	}
	args := strings.Split(m.Message.Content, " ")
	mention := "<@" + *appID + ">"
	if args[0] != mention {
		return
	}
	help := mention + " is a simple discord bot for creating private channels for you and your friends." +
		"This channel will be automatically deleted when you leave it.\n\n" +
		"Usage:\n" +
		mention + " `<size of channel>`. To increase user limit just send new size.\n\n"
	if len(args) < 2 {
		s.ChannelMessageSendEmbed(m.Message.ChannelID, &discordgo.MessageEmbed{
			Description: help,
		})
		return
	}

	size, err := strconv.Atoi(args[1])
	if err != nil {
		s.ChannelMessageSend(m.Message.ChannelID, "Size can only be integer")
		return
	}
	if size < 2 || size > 100 {
		s.ChannelMessageSend(m.Message.ChannelID, "Choose number between 2 and 100")
		return
	}
	err = createChannel(s, channel, m.Message.Author, size)
	if err != nil {
		fmt.Println(err)
		s.ChannelMessageSend(channel.ID, "Something went wrong creating (editing) your channel")
	}
}

func createChannel(s *discordgo.Session, c *discordgo.Channel, owner *discordgo.User, size int) error {
	elem, ok := rented[owner.ID]
	if ok {
		channel, err := s.Channel(elem.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		_, err = s.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{
			UserLimit: size,
			Position:  channel.Position,
		})
		if err != nil {
			fmt.Println(err)
			return err
		}
		s.ChannelMessageSend(c.ID, "Edited your rented channel")
		return nil
	}
	channel, err := s.GuildChannelCreateComplex(c.GuildID, discordgo.GuildChannelCreateData{
		Name:      owner.Username + "'s Channel",
		Type:      discordgo.ChannelTypeGuildVoice,
		UserLimit: size,
	})
	if err != nil {

		return err
	}
	s.ChannelMessageSend(c.ID, "Created channel for you")
	rented[owner.ID] = &rentChannel{
		owner:   owner.ID,
		ID:      channel.ID,
		visited: false,
	}
	time.AfterFunc(20*time.Second, func() {
		if !rented[owner.ID].visited {
			s.ChannelDelete(rented[owner.ID].ID)
		}
	})
	return nil
}
