package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	// all the emojis we use
	leftArrow, rightArrow, destroyEmoji = "⬅️", "➡️", "❌"
)

var (
	pageListeners = make(map[string]*ReactionListener)
)

// MakeTimestamp - makes a ms timestamp
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func checkListeners() {
	for {
		time.Sleep((60 * 5) * time.Second)
		var now = MakeTimestamp()
		for key, listener := range pageListeners {
			if (now - listener.LastUsed) > 300000 {
				delete(pageListeners, key)
			}
		}
	}
}

// here is where we listen for reactions (for the pages)
func reactionListen(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {

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
			// update last used so the lsitener isn't deemed inactive
			pageListeners[reaction.MessageID].LastUsed = MakeTimestamp()
			// remove reaction, better user expirence
			session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, leftArrow, reaction.UserID)
			// If page is already 1 aka lowest it can be
			if pageListeners[reaction.MessageID].CurrentPage == 1 {
				break
			}
			// decrease current page
			pageListeners[reaction.MessageID].CurrentPage--
			textTosend := formatForMessage(pageListeners[reaction.MessageID])
			session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
				Title:       pageListeners[reaction.MessageID].Type,
				Description: textTosend,
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
			textTosend := formatForMessage(pageListeners[reaction.MessageID])
			session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
				Title:       pageListeners[reaction.MessageID].Type,
				Description: textTosend,
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
