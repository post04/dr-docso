package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// New creates an initialized commandhandler
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

// GenHelp generates the help command output.
func (handler *CommandHandler) GenHelp() {
	handler.TimeStarted = time.Now()

	var longestCommand int
	for name := range handler.Commands {
		if len(name) > longestCommand {
			longestCommand = len(name)
		}
	}

	// Build the description
	embedDesc := strings.Builder{}
	embedDesc.WriteString("```autoit\n")
	for name := range handler.Commands {
		embedDesc.WriteString(handler.Prefix)
		embedDesc.WriteString(name)
		embedDesc.WriteString(strings.Repeat(" ", (longestCommand-len(name))+1))
		embedDesc.WriteString(handler.Commands[name].Description)
		embedDesc.WriteRune('\n')
	}
	embedDesc.WriteString("```")

	handler.HelpCommand = &discordgo.MessageEmbed{
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Use %shelp commandName to get more info about a command", handler.Prefix),
		},
		Title:       "Available commands",
		Description: embedDesc.String(),
	}
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

	if handler.Middleware != nil {
		if !handler.Middleware(session, msg) {
			return
		}
	}
	cmd := strings.ToLower(strings.TrimPrefix(parts[0], handler.Prefix))

	switch cmd {
	case "help":
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
				embed.Title = strings.ToLower(parts[1]) + " Command"
				session.ChannelMessageSendEmbed(msg.ChannelID, embed)
			} else {
				embed.Title = "Unknown command"
				embed.Description = fmt.Sprintf("%q is not a valid command; use %shelp to get available commands.", parts[1], handler.Prefix)
				session.ChannelMessageSendEmbed(msg.ChannelID, embed)
			}
		}
	case "info":
		fmt.Println("info command ran by " + msg.Author.Username + "#" + msg.Author.Discriminator + " in " + msg.ChannelID)

		session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
			Title:       "dr-docso by post and insomnia",
			Description: fmt.Sprintf("Library: [DiscordGo](https://github.com/bwmarrin/discordgo)\nUptime: %s\nPrefix: %s\nGithub Repo: [here](https://github.com/post04/dr-docso)\nInvite: [Click me](https://discord.com/oauth2/authorize?client_id=817416218390560798&permissions=3221613648&scope=bot)", time.Since(handler.TimeStarted), handler.Prefix),
		})
	default:
		if command, ok := handler.Commands[cmd]; ok {
			fmt.Println(parts[0][len(handler.Prefix):] + " command ran by " + msg.Author.Username + "#" + msg.Author.Discriminator + " in " + msg.ChannelID)
			go command.Run(session, msg, handler.Prefix)
		}
	}
}
