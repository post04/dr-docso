package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/post04/dr-docso/docs"
)

const (
	// all the emojis we use
	leftArrow    = "⬅️"
	rightArrow   = "➡️"
	destroyEmoji = "❌"
)

// ReactionListener is a struct for the reaction listener for pages
type ReactionListener struct {
	Type        string
	CurrentPage int
	PageLimit   int
	UserID      string
	Data        *docs.Doc
	LastUsed    int64 // use uint64 maybe?
}

var (
	pageListeners = make(map[string]*ReactionListener)
)

// MakeTimestamp makes an ms timestamp
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// CheckListeners checks all active reaction listeners and kills inactive ones
func CheckListeners() {
	for {
		time.Sleep(6 * time.Minute)
		for key, listener := range pageListeners {
			if MakeTimestamp()-listener.LastUsed > 300000 {
				delete(pageListeners, key)
			}
		}
	}
}

func formatForMessage(page *ReactionListener) string {
	s := ""
	max := page.CurrentPage * 10
	min := max - 10
	curr := min
	if page.Type == "functions" {
		if max > len(page.Data.Functions) {
			max = len(page.Data.Functions)
		}
		for _, function := range page.Data.Functions[min:max] {
			curr++
			s += fmt.Sprintf("\n%v.) %v", curr, function.Name)
		}
	}
	if page.Type == "types" {
		if max > len(page.Data.Types) {
			max = len(page.Data.Types)
		}
		for _, dType := range page.Data.Types[min:max] {
			curr++
			s += fmt.Sprintf("\n%v.) %v", curr, dType.Name)
		}
	}
	return s
}

// ReactionListen listens for the reactions for a previously sent embed.
func ReactionListen(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	// if the message being reacted to is in the reaction map
	if _, ok := pageListeners[reaction.MessageID]; ok {
		// validating that the user reacting is indeed the user that owns the listener
		if pageListeners[reaction.MessageID].UserID != reaction.UserID {
			return
		}
		// switch the emoji name, name is misleading, basically -> ❌
		switch reaction.Emoji.Name {
		// when the reaction used is a left arrow (page decrease)
		case leftArrow:
			// update last used so the listener isn't deemed inactive
			pageListeners[reaction.MessageID].LastUsed = MakeTimestamp()
			// remove reaction, better user experience
			session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, leftArrow, reaction.UserID)
			// If page is already 1 (minimum it can be)
			if pageListeners[reaction.MessageID].CurrentPage == 1 {
				break
			}
			// decrease current page
			pageListeners[reaction.MessageID].CurrentPage -= 1
			session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
				Title:       pageListeners[reaction.MessageID].Type,
				Description: formatForMessage(pageListeners[reaction.MessageID]),
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Page %v/%v", pageListeners[reaction.MessageID].CurrentPage, pageListeners[reaction.MessageID].PageLimit),
				},
			})
			break
		case rightArrow:
			// update last used so the listener isn't deemed unused and deleted
			pageListeners[reaction.MessageID].LastUsed = MakeTimestamp()
			// remove reaction for better user expirence
			session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, rightArrow, reaction.UserID)
			// if the page we're on right now is already the maximum page length we have, on page 7 out of 7
			if pageListeners[reaction.MessageID].PageLimit == pageListeners[reaction.MessageID].CurrentPage {
				break
			}
			// update current page by 1
			pageListeners[reaction.MessageID].CurrentPage++
			session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
				Title:       pageListeners[reaction.MessageID].Type,
				Description: formatForMessage(pageListeners[reaction.MessageID]),
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Page %v/%v", pageListeners[reaction.MessageID].CurrentPage, pageListeners[reaction.MessageID].PageLimit),
				},
			})
			// done :sunglasses:
			break
		case destroyEmoji:
			// remove the specific page listener from the map, no longer listening for reactions
			delete(pageListeners, reaction.MessageID)
			// delete the embed the bot made, just cleans itself up.
			session.ChannelMessageDelete(reaction.ChannelID, reaction.MessageID)
			// done :sunglasses:
			break
		default:
			// done :sunglasses:
			break
		}
	}
}
