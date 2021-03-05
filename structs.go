package main

import "github.com/bwmarrin/discordgo"

type Command struct {
	Help        string
	Description string
	Run         func(session *discordgo.Session, msg *discordgo.MessageCreate, prefix string)
}

type CommandHandler struct {
	Prefix            string
	Commands          map[string]*Command
	IgnoreBots        bool
	OnMessageHandler  func(session *discordgo.Session, msg *discordgo.MessageCreate)
	PreCommandHandler func(session *discordgo.Session, msg *discordgo.MessageCreate) bool
	HelpCommand       *discordgo.MessageEmbed
}

type Config struct {
	Prefix         string   `json:"prefix"`
	Token          string   `json:"token"`
	MainGuild      string   `json:"mainGuild"`
	LockedChannels []string `json:"lockedChannels"`
	SafeMode       bool     `json:"safeMode"`
}
