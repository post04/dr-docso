package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/postrequest69/dr-docso/docs"
)

// Command - Command struct
type Command struct {
	Help        string
	Description string
	Run         func(session *discordgo.Session, msg *discordgo.MessageCreate, args []string, prefix string)
}

// CommandHandler lol
type CommandHandler struct {
	Prefix            string
	Commands          map[string]*Command
	IgnoreBots        bool
	OnMessageHandler  func(session *discordgo.Session, msg *discordgo.MessageCreate)
	PreCommandHandler func(session *discordgo.Session, msg *discordgo.MessageCreate) bool
}

// Config config struct
type Config struct {
	Prefix         string   `json:"prefix"`
	Token          string   `json:"token"`
	MainGuild      string   `json:"mainGuild"`
	LockedChannels []string `json:"lockedChannels"`
	SafeMode       bool     `json:"safeMode"`
}

// ReactionListener is a struct for the reaction listener for pages
type ReactionListener struct {
	Type        string
	CurrentPage int
	PageLimit   int
	UserID      string
	Data        *docs.Doc
	LastUsed    int64
}
