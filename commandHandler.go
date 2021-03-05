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
		Commands:   make(map[string]*Command),
		IgnoreBots: ignoreBots,
	}
}

// AddCommand adds a new command to command handler
func (handler *CommandHandler) AddCommand(name string, help string, description string, commandHandler func(s *discordgo.Session, m *discordgo.MessageCreate, prefix string)) {
	help = strings.ReplaceAll(help, "{prefix}", handler.Prefix)
	handler.Commands[name] = &Command{
		Run:         commandHandler,
		Help:        help,
		Description: description,
	}
}

// GenHelp makes the help command embed so it wont be generated every time the command is used
func (handler *CommandHandler) GenHelp() {
	embed := &discordgo.MessageEmbed{
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Reminder: You can define a command to get help with like %s help commandName", handler.Prefix),
		},
		Title:       "Showing all possible commands!",
		Description: "```autoit\n",
	}

	var longestCommand int
	desc := []string{}
	for name := range handler.Commands {
		if len(name) > longestCommand {
			longestCommand = len(name)
		}
		desc = append(desc, name)
	}
	for pos, descPiece := range desc {
		spaces := " "
		if len(descPiece) < longestCommand {
			spaces = strings.Repeat(" ", longestCommand-len(descPiece)+1)
		}
		desc[pos] = fmt.Sprintf("%v%v%v#%v", handler.Prefix, descPiece, spaces, handler.Commands[descPiece].Description)
	}
	embed.Description += strings.Join(desc, "\n") + "```"
	handler.HelpCommand = embed
}

// OnMessage handles onmessage event from discordgo for command handler lol
func (handler *CommandHandler) OnMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	parts := strings.Fields(msg.Content)
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

			session.ChannelMessageSendEmbed(msg.ChannelID, handler.HelpCommand)
		} else {
			embed := &discordgo.MessageEmbed{}
			if command, ok := handler.Commands[strings.ToLower(parts[1])]; ok {
				embed.Description = fmt.Sprintf("Name: %s\n"+
					"example: %s\n"+
					"description: %s",
					strings.ToLower(parts[1]),
					command.Help,
					command.Description)
				embed.Title = strings.ToLower(parts[1]) + " Command!"
				session.ChannelMessageSendEmbed(msg.ChannelID, embed)
			} else {
				embed.Title = "Invalid command!"
				embed.Description = fmt.Sprintf("Please use %s help to get all commands!", handler.Prefix)
				session.ChannelMessageSendEmbed(msg.ChannelID, embed)
			}
		}
	} else {
		if command, ok := handler.Commands[strings.ToLower(parts[0][len(handler.Prefix):])]; ok {
			fmt.Println(parts[0][len(handler.Prefix):] + " command ran by " + msg.Author.Username + "#" + msg.Author.Discriminator + " in " + msg.ChannelID)
			go command.Run(session, msg, handler.Prefix)
		}
	}
}
