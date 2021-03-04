package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	bot "github.com/postrequest69/dr-docso/bot"
	"github.com/postrequest69/dr-docso/docs"
)

// DocsCommand is the discord command for viewing documentation from discord!
func DocsCommand(session *discordgo.Session, msg *discordgo.MessageCreate, arguments []string, prefix string) {
	if len(arguments) < 1 {
		session.ChannelMessageSendEmbed(msg.ChannelID, bot.DocsHelpEmbed)
		return
	}

	// if the command is being used to just get information on the package
	if len(arguments) == 1 {
		doc, err := docs.GetDoc(arguments[0])
		if err != nil {
			session.ChannelMessageSend(msg.ChannelID, "An error occured while trying to get this package!")
			return
		}
		if len(doc.Types) == 0 && len(doc.Functions) == 0 {
			session.ChannelMessageSend(msg.ChannelID, "It seems this package doesn't exist according to pkg.go.dev!")
			return
		}
		session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Info for %s", arguments[0]),
			Description: fmt.Sprintf("Types: %d\nFunctions: %d", len(doc.Types), len(doc.Functions)),
		})
		return
	}

	// if the command is being used to get a list of the functions or types of a package
	if len(arguments) == 2 {
		doc, err := docs.GetDoc(arguments[0])
		if err != nil {
			session.ChannelMessageSend(msg.ChannelID, "An error occured while trying to get this package!")
			return
		}
		if len(doc.Types) == 0 && len(doc.Functions) == 0 {
			session.ChannelMessageSend(msg.ChannelID, "It seems this package doesn't exist according to pkg.go.dev!")
			return
		}
		name := arguments[1]
		var s string
		for _, function := range doc.Functions {
			if strings.EqualFold(function.Name, name) {
				s += fmt.Sprintf("`%s`", function.Signature)
				if len(function.Comments) > 0 {
					s += fmt.Sprintf("\n%s", function.Comments[0])
				} else {
					s += "\nNo comment available"
				}
				if function.Example != "" {
					s += fmt.Sprintf("\nExample:\n```go\n%s\n```", function.Example)
				}
			}
		}
		if s == "" {
			s = "no information available"
		} else if len(s) > 2000 {
			s = fmt.Sprintf("%s\n*message trimmed to fit 2k char limit*", s[:1800])
		}
		session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
			Title:       "Function: " + name,
			Description: s,
		})
		return
	}

	// if the command is being used to get a specific function or type of a package
	if len(arguments) == 3 {
		arg := strings.ToLower(arguments[1])
		if arg != "function" && arg != "type" {
			session.ChannelMessageSendEmbed(msg.ChannelID, bot.DocsHelpEmbed)
			return
		}
		doc, err := docs.GetDoc(arguments[0])
		if err != nil {
			session.ChannelMessageSend(msg.ChannelID, "An error occured while trying to get this package!")
			return
		}
		if len(doc.Types) == 0 && len(doc.Functions) == 0 {
			session.ChannelMessageSend(msg.ChannelID, "It seems this package doesn't exist according to pkg.go.dev!")
			return
		}
		if arg == "function" {
			name := arguments[2]
			var s string
			for _, function := range doc.Functions {
				if strings.EqualFold(function.Name, name) {
					s += fmt.Sprintf("`%s`", function.Signature)
					if len(function.Comments) > 0 {
						s += fmt.Sprintf("\n%s", function.Comments[0])
					} else {
						s += "\nNo comment available"
					}
					if function.Example != "" {
						s += fmt.Sprintf("\nExample:\n```go\n%s\n```", function.Example)
					}
				}
			}
			if s == "" {
				s = "no information available"
			} else if len(s) > 2000 {
				s = fmt.Sprintf("%s\n*message trimmed to fit 2k char limit*", s[:1800])
			}
			session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
				Title:       "Function: " + name,
				Description: s,
			})
			return
		}
		if arg == "type" {
			name := arguments[2]
			var s string
			for _, dType := range doc.Types {
				if strings.EqualFold(dType.Name, name) {
					s += fmt.Sprintf("```go\n%s\n```", dType.Signature)
					if len(dType.Comments) > 0 {
						s += fmt.Sprintf("\n%s", dType.Comments[0])
					} else {
						s += "no comments available"
					}
				}
			}
			if s == "" {
				s = "no information available"
			} else if len(s) > 2000 {
				s = fmt.Sprintf("%s\n*message trimmed to fit the 2k char limit*", s[:1900])
			}
			session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
				Title:       "Type: " + name,
				Description: s,
			})
			return
		}
	}
	if len(arguments) > 3 {
		session.ChannelMessageSendEmbed(msg.ChannelID, bot.DocsHelpEmbed)
		return
	}
}
