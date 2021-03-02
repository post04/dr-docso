package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/postrequest69/dr-docso/docs"
)

func formatForMessage(page *ReactionListener) string {
	toReturn := ""
	var max = page.CurrentPage * 10
	max--
	var min = max - 9
	curr := min
	if page.Type == "functions" {
		if max > len(page.Data.Functions) {
			max = len(page.Data.Functions)
		}
		for _, function := range page.Data.Functions[min:max] {
			curr++
			toReturn += fmt.Sprintf("\n%v.) %v", curr, function.Name)
		}
	}
	if page.Type == "types" {
		if max > len(page.Data.Types) {
			max = len(page.Data.Types)
		}
		for _, dType := range page.Data.Types[min:max] {
			curr++
			toReturn += fmt.Sprintf("\n%v.) %v", curr, dType.Name)
		}
	}
	return toReturn
}

// DocsCommand is the discord command for viewing documentation from discord!
func DocsCommand(session *discordgo.Session, msg *discordgo.MessageCreate, arguments []string, prefix string) {
	if len(arguments) < 1 {
		session.ChannelMessageSendEmbed(msg.ChannelID, docsCommandHelpEmbed)
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
			Title:       "Info for " + arguments[0],
			Description: fmt.Sprintf("Types: %v\nFunctions: %v", len(doc.Types), len(doc.Functions)),
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
		// THIS IS FOR REACTION PAGES COMMAND BTW LOL
		// if arg == "functions" {
		// 	var pageLimit = int(math.Ceil(float64(len(doc.Functions)) / 10.0))
		// 	var page = &ReactionListener{
		// 		Type:        "functions",
		// 		CurrentPage: 1,
		// 		PageLimit:   pageLimit,
		// 		UserID:      msg.Author.ID,
		// 		Data:        doc,
		// 		LastUsed:    MakeTimestamp(),
		// 	}
		// 	textTosend := formatForMessage(page)
		// 	m, err := session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
		// 		Title:       "functions",
		// 		Description: textTosend,
		// 		Footer: &discordgo.MessageEmbedFooter{
		// 			Text: "Page 1/" + fmt.Sprint(pageLimit),
		// 		},
		// 	})
		// 	if err != nil {
		// 		return
		// 	}
		// 	session.MessageReactionAdd(m.ChannelID, m.ID, leftArrow)
		// 	session.MessageReactionAdd(m.ChannelID, m.ID, rightArrow)
		// 	session.MessageReactionAdd(m.ChannelID, m.ID, destroyEmoji)
		// 	pageListeners[m.ID] = page
		// 	return
		// }
		// if arg == "types" {
		// 	var pageLimit = int(math.Ceil(float64(len(doc.Types)) / 10.0))
		// 	var page = &ReactionListener{
		// 		Type:        "types",
		// 		CurrentPage: 1,
		// 		PageLimit:   pageLimit,
		// 		UserID:      msg.Author.ID,
		// 		Data:        doc,
		// 		LastUsed:    MakeTimestamp(),
		// 	}
		// 	textTosend := formatForMessage(page)
		// 	m, err := session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
		// 		Title:       "types",
		// 		Description: textTosend,
		// 		Footer: &discordgo.MessageEmbedFooter{
		// 			Text: "Page 1/" + fmt.Sprint(pageLimit),
		// 		},
		// 	})
		// 	if err != nil {
		// 		return
		// 	}
		// 	session.MessageReactionAdd(m.ChannelID, m.ID, leftArrow)
		// 	session.MessageReactionAdd(m.ChannelID, m.ID, rightArrow)
		// 	session.MessageReactionAdd(m.ChannelID, m.ID, destroyEmoji)
		// 	pageListeners[m.ID] = page
		// 	return
		// }
		name := arguments[1]
		var toSend string
		for _, function := range doc.Functions {
			if strings.EqualFold(function.Name, name) {
				if function.Example != "" {
					function.Example = fmt.Sprintf("```go\n%v```", function.Example)
				}
				if function.Example == "" {
					function.Example = "None"
				}
				if function.MethodOf == "" {
					function.MethodOf = "None"
				}

				if len(function.Comments) == 0 {
					function.Comments = append(function.Comments, "None")
				}
				toSend += fmt.Sprintf("%v\nType: %v\nMethodOf: %v\nExample: %v\nComments: ```%v```\n", function.Signature, function.Type, function.MethodOf, function.Example, function.Comments[0])
			}
		}
		if toSend == "" {
			toSend = "No information avalible!"
		}
		session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
			Title:       "Function: " + name,
			Description: toSend,
		})
		return
	}
	// if the command is being used to get a specific function or type of a package
	if len(arguments) == 3 {
		arg := strings.ToLower(arguments[1])
		if arg != "functions" && arg != "types" {
			session.ChannelMessageSendEmbed(msg.ChannelID, docsCommandHelpEmbed)
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
		if arg == "functions" {
			name := arguments[2]
			var toSend string
			for _, function := range doc.Functions {
				if strings.EqualFold(function.Name, name) {
					if function.Example == "" {
						function.Example = "None"
					}
					if len(function.Comments) == 0 {
						function.Comments = append(function.Comments, "None")
					}
					toSend += fmt.Sprintf("%v\nType: %v\nMethodOf: %v\nExample: %v\nComments: ```%v```\n", function.Signature, function.Type, function.MethodOf, function.Example, function.Comments[0])
				}
			}
			session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
				Title:       "Function: " + name,
				Description: toSend,
			})
			return
		}
		if arg == "types" {
			name := arguments[2]
			var toSend string
			for _, dType := range doc.Types {
				if strings.EqualFold(dType.Name, name) {
					if len(dType.Comments) == 0 {
						dType.Comments = append(dType.Comments, "None")
					}
					toSend += fmt.Sprintf("Type: %v\nSignature: ```go\n%v```\nComments: %v\n", dType.Type, dType.Signature, dType.Comments[0])
				}
			}
			session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
				Title:       "Type: " + name,
				Description: toSend,
			})
			return
		}
	}
	if len(arguments) > 3 {
		session.ChannelMessageSendEmbed(msg.ChannelID, docsCommandHelpEmbed)
		return
	}
}
