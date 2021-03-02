package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// New Registers a new commandhandler
func New(prefix string, ignoreBots bool) *CommandHandler {
	return &CommandHandler{
		Prefix:     prefix,
		Commands:   map[string]*Command{},
		IgnoreBots: ignoreBots,
	}
}

// AddCommand adds a new command to command handler
func (handler *CommandHandler) AddCommand(name string, help string, description string, commandHandler func(s *discordgo.Session, m *discordgo.MessageCreate, args []string, prefix string)) {
	help = strings.ReplaceAll(help, "{prefix}", handler.Prefix)
	handler.Commands[name] = &Command{
		Run:         commandHandler,
		Help:        help,
		Description: description,
	}
}

// OnMessage handles onmessage event from discordgo for command handler lol
func (handler *CommandHandler) OnMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	parts := strings.Split(msg.Content, " ")
	if handler.OnMessageHandler != nil {
		go handler.OnMessageHandler(session, msg)
	}
	if msg.Author.Bot && handler.IgnoreBots {
		return
	}
	if len(parts) < 1 || !strings.HasPrefix(parts[0], handler.Prefix) {
		return
	}
	if handler.PreCommandHandler != nil {
		if !handler.PreCommandHandler(session, msg) {
			return
		}
	}
	if strings.ToLower(parts[0][len(handler.Prefix):]) == "help" {
		fmt.Println("help command ran by " + msg.Author.Username + "#" + msg.Author.Discriminator + " in " + msg.ChannelID)
		if len(parts) == 1 {
			var embed = &discordgo.MessageEmbed{
				Footer:      &discordgo.MessageEmbedFooter{},
				Description: "",
				Title:       "",
			}
			var foot = "Reminder: You can define a command to get help with like " + handler.Prefix + "help commandName"
			embed.Footer.Text = foot
			var desc string
			embed.Title = "Showing all possible commands!"
			var i = 0
			for name := range handler.Commands {
				i++
				desc += fmt.Sprint(i) + ".) " + name + "\n"
			}
			embed.Description = desc
			session.ChannelMessageSendEmbed(msg.ChannelID, embed)
		} else {
			var embed = &discordgo.MessageEmbed{
				Description: "",
				Title:       "",
			}
			if command, ok := handler.Commands[strings.ToLower(parts[1])]; ok {
				var desc string
				desc += "Name: " + strings.ToLower(parts[1]) + "\n"
				desc += "Help: " + command.Help
				desc += "\n" + "Description: " + command.Description
				embed.Description = desc
				embed.Title = strings.ToLower(parts[1]) + " Command!"
				session.ChannelMessageSendEmbed(msg.ChannelID, embed)
			} else {
				embed.Title = "Invalid command!"
				embed.Description = "Please use " + handler.Prefix + "help to get all commands!"
				session.ChannelMessageSendEmbed(msg.ChannelID, embed)
			}
		}
	} else {
		if command, ok := handler.Commands[strings.ToLower(parts[0][len(handler.Prefix):])]; ok {
			fmt.Println(parts[0][len(handler.Prefix):] + " command ran by " + msg.Author.Username + "#" + msg.Author.Discriminator + " in " + msg.ChannelID)
			go command.Run(session, msg, parts[1:], handler.Prefix)
		}
	}
}
