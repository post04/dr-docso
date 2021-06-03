package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/post04/dr-docso/docs"
	"github.com/post04/dr-docso/glob"
)

const regexpSpecials = "*|[]()+{}-"

// DocsHelpEmbed is the help for the docs command.
var DocsHelpEmbed = &discordgo.MessageEmbed{
	Title: "Docs help",
}

// HandleDoc  is the handler for the doc command.
func HandleDoc(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	var msg *discordgo.MessageEmbed
	fields := strings.Fields(m.Content)
	switch len(fields) {
	case 0: // probably impossible
		return
	case 1: // only the invocation
		msg = helpShortResponse() // TODO: probably should just use the variable here.
	case 2: // invocation + arg
		msg = pkgResponse(fields[1])
	case 3: // invocation + pkg + func
		if strings.Contains(fields[2], ".") {
			split := strings.Split(fields[2], ".")
			msg = methodResponse(fields[1], split[0], split[1])
		} else {
			msg = queryResponse(fields[1], fields[2])
		}
	default:
		msg = errResponse("Too many arguments.")
	}

	if msg == nil {
		msg = errResponse("No results found, possibly an internal error.")
	}
	embedM, err := s.ChannelMessageSendEmbed(m.ChannelID, msg)
	if err != nil {
		log.Printf("could not send message: %s", err)
	}

	err = s.MessageReactionAdd(embedM.ChannelID, embedM.ID, destroyEmoji)
	if err != nil {
		log.Printf("could not send message: %s", err)
	}
}

// queryResponse generates the response for a query.
//
// i.e, `.docs strings Builder`
func queryResponse(pkg, name string) *discordgo.MessageEmbed {
	if strings.ContainsAny(name, regexpSpecials) {
		return queryGlobResponse(pkg, name)
	}
	doc, err := getDoc(pkg)
	if err != nil {
		return errResponse("An error occurred while fetching the page for pkg `%s`", pkg)
	}
	var msg string
	for _, fn := range doc.Functions {
		// Why is the first check there? // because this is not a method query.
		if fn.Type == docs.FnNormal && strings.EqualFold(fn.Name, name) {
			// match found
			name = fn.Name
			msg += fmt.Sprintf("`%s`", fn.Signature)
			if len(fn.Comments) > 0 {
				msg += fmt.Sprintf("\n%s", strings.Join(fn.Comments, "\n"))
			} else {
				msg += "\n*no information*"
			}
			if fn.Example != "" {
				msg += fmt.Sprintf("\n\nExample:\n```go\n%s\n```", fn.Example)
			}
		}
	}

	if msg == "" {
		for _, t := range doc.Types {
			if strings.EqualFold(name, t.Name) {
				msg += fmt.Sprintf("```go\n%s\n```\n", t.Signature)
				if len(t.Comments) > 0 {
					msg += strings.Join(t.Comments, "\n")
				} else {
					msg += "*no information available*\n"
				}
			}
		}
	}

	if msg == "" {
		return errResponse("No type or function `%s` found in package `%s`", name, pkg)
	}
	if len(msg) > 2000 {
		msg = fmt.Sprintf("%s\n\n*note: the message was trimmed to fit the 2k character limit*", msg[:1950])
	}
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s: %s", pkg, name),
		URL:         fmt.Sprintf("%s#%s", doc.URL, correctName(name)),
		Description: msg,
		Footer: &discordgo.MessageEmbedFooter{
			// NOTE(insomnia) please do not use %s when formatting strings, it's not for strings, it's for GoStringer.
			Text: fmt.Sprintf("%s#%s", doc.URL, correctName(name)),
		},
	}
}

// this function is used to capitalize the first letter of a word
func correctName(word string) string {
	first := word[:1]
	word = word[1:]
	word = strings.ToUpper(first) + word
	return word
}

// errResponse is like fmt.Sprintf, formats a message and returns an embed.
func errResponse(format string, args ...interface{}) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: fmt.Sprintf(format, args...),
	}
}

// helpShortResponse returns the docs command's help embed.
func helpShortResponse() *discordgo.MessageEmbed {
	return DocsHelpEmbed
}

// pkgResponse generates an embed with general information about a package.
func pkgResponse(pkg string) *discordgo.MessageEmbed {
	doc, err := getDoc(pkg)
	if err != nil {
		return errResponse("An error occured when requesting the page for the package `%s`", pkg)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Info for %s", pkg),
		URL:         fmt.Sprintf("%s", doc.URL),
		Description: fmt.Sprintf("Types: %v\nFunctions: %v", len(doc.Types), len(doc.Functions)),
		Footer: &discordgo.MessageEmbedFooter{
			Text: doc.URL,
		},
	}
	if doc.Overview != "" {
		embed.Description += fmt.Sprintf("\nOverview: %s", doc.Overview)
	}
	if len(embed.Description) > 2000 {
		embed.Description = embed.Description[:1900] + "\n*Note this embed has been cut because it is too long*"
	}
	return embed
}

// methodResponse generates an embed for a method query.
//
// i.e, `.docs regexp Regexp.Match`
func methodResponse(pkg, t, name string) *discordgo.MessageEmbed {
	if strings.ContainsAny(t, regexpSpecials) ||
		strings.ContainsAny(name, regexpSpecials) {
		return methodGlobResponse(pkg, t, name)
	}

	doc, err := getDoc(pkg)
	if err != nil {
		return errResponse("Error while getting the page for the package `%s`", pkg)
	}
	if len(doc.Functions) == 0 {
		return errResponse("Package `%s` seems to have no functions", pkg)
	}

	var msg string
	var hyper string
	for _, fn := range doc.Functions {
		if fn.Type == docs.FnMethod &&
			strings.EqualFold(fn.Name, name) &&
			strings.EqualFold(fn.MethodOf, t) {
			hyper = fmt.Sprintf("%s#%s.%s", doc.URL, fn.MethodOf, fn.Name)
			msg += fmt.Sprintf("`%s`", fn.Signature)
			if len(fn.Comments) > 0 {
				msg += fmt.Sprintf("\n%s", fn.Comments[0])
			} else {
				msg += "\n*no info*"
			}
			if fn.Example != "" {
				msg += fmt.Sprintf("\nExample:\n```go\n%s\n```", fn.Example)
			}
		}
	}
	if msg == "" {
		return errResponse("Package `%s` does not have `func(%s) %s`", pkg, t, name)
	}
	if len(msg) > 2000 {
		msg = fmt.Sprintf("%s\n\n*note: the message is trimmed to fit the 2k character limit*", msg[:1950])
	}
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s: func(%s) %s", pkg, t, name),
		URL:         hyper,
		Description: msg,
		Footer: &discordgo.MessageEmbedFooter{
			Text: hyper,
		},
	}
}

func PagesShortResponse(state, prefix string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Help %s", state),
		Description: fmt.Sprintf("It seems you didn't have enough arguments, so here's an example\n\n%s%s strings", prefix, state),
	}
}

// HandleFuncsPages is the handler fo the getfuncs command
func HandleFuncsPages(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	fields := strings.Fields(m.Content)
	switch len(fields) {
	case 0: // probably impossible
		return
	case 1: // send a help command here
		s.ChannelMessageSendEmbed(m.ChannelID, PagesShortResponse("getfuncs", prefix))
		return
	case 2: // command + pkg (send page if possible)
		doc, err := getDoc(fields[1])
		if err != nil || doc == nil {
			s.ChannelMessageSendEmbed(m.ChannelID, errResponse("Error while getting the page for the package `%s`", fields[1]))
			return
		}
		if len(doc.Functions) == 0 {
			s.ChannelMessageSendEmbed(m.ChannelID, errResponse("The package `%s` has no functions", fields[1]))
			return
		}
		page := &ReactionListener{
			Type:        "functions",
			CurrentPage: 1,
			PageLimit:   calcLimit(len(doc.Functions), 10),
			UserID:      m.Author.ID,
			Data:        doc,
			LastUsed:    time.Now(),
		}
		m, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       "functions",
			URL:         doc.URL + "#pkg-functions",
			Description: formatForMessage(page),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Page 1/" + fmt.Sprint(page.PageLimit),
			},
		})
		if err != nil {
			return
		}
		s.MessageReactionAdd(m.ChannelID, m.ID, leftArrow)
		s.MessageReactionAdd(m.ChannelID, m.ID, rightArrow)
		s.MessageReactionAdd(m.ChannelID, m.ID, destroyEmoji)
		pageListeners[m.ID] = page
		return
	default: // too many arguments
		s.ChannelMessageSendEmbed(m.ChannelID, PagesShortResponse("getfuncs", prefix))
		return
	}

}

// HandleTypesPages is the handler fo the gettypes command
func HandleTypesPages(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	fields := strings.Fields(m.Content)
	switch len(fields) {
	case 0: // probably impossible
		return
	case 1: // send a help command here
		s.ChannelMessageSendEmbed(m.ChannelID, PagesShortResponse("gettypes", prefix))
		return
	case 2: // command + pkg (send page if possible)
		doc, err := getDoc(fields[1])
		if err != nil || doc == nil {
			s.ChannelMessageSendEmbed(m.ChannelID, errResponse("Error while getting the page for the package `%s`", fields[1]))
			return
		}
		if len(doc.Functions) == 0 {
			s.ChannelMessageSendEmbed(m.ChannelID, errResponse("The package `%s` has no types", fields[1]))
			return
		}
		page := &ReactionListener{
			Type:        "types",
			CurrentPage: 1,
			PageLimit:   calcLimit(len(doc.Types), 10),
			UserID:      m.Author.ID,
			Data:        doc,
			LastUsed:    time.Now(),
		}
		m, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       "types",
			URL:         doc.URL + "#pkg-types",
			Description: formatForMessage(page),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Page 1/" + fmt.Sprint(page.PageLimit),
			},
		})
		if err != nil {
			return
		}
		s.MessageReactionAdd(m.ChannelID, m.ID, leftArrow)
		s.MessageReactionAdd(m.ChannelID, m.ID, rightArrow)
		s.MessageReactionAdd(m.ChannelID, m.ID, destroyEmoji)
		pageListeners[m.ID] = page
		return
	default: // too many arguments
		s.ChannelMessageSendEmbed(m.ChannelID, PagesShortResponse("gettypes", prefix))
		return
	}
}

// methodGlobResponse generates an embed for a glob pattern describing type.method.
func methodGlobResponse(pkg, t, name string) *discordgo.MessageEmbed {
	reT, err := glob.Compile(t)
	if err != nil {
		return errResponse("Error processing glob pattern:\n```\n%s\n```", err)
	}
	reN, err := glob.Compile(name)
	if err != nil {
		return errResponse("Error processing glob pattern:\n```\n%s\n```", err)
	}
	doc, err := getDoc(pkg)
	if err != nil {
		return errResponse("An error occurred while getting the page for the package `%s`", pkg)
	}

	if len(doc.Functions) == 0 || len(doc.Types) == 0 {
		return errResponse("No results found matching the expression `%s.%s` in package `%s`", t, name, pkg)
	}

	var msg string
	for _, fn := range doc.Functions {
		if fn.Type == docs.FnMethod &&
			reT.MatchString(fn.MethodOf) &&
			reN.MatchString(fn.Name) {
			msg += fmt.Sprintf("`%s`:\n", fn.Signature)
			if len(fn.Comments) > 0 {
				msg += fn.Comments[0]
			} else {
				msg += "*no information available*"
			}
		}
	}
	if msg == "" {
		return errResponse("No results found matching the expression `%s.%s` in package `%s`", t, name, pkg)
	}
	if len(msg) > 2000 {
		msg = fmt.Sprintf("%s\n\n*note: the message was trimmed to fit the 2k character limit*", msg[:1950])
	}
	return &discordgo.MessageEmbed{
		Title:       "Matches",
		URL:         doc.URL,
		Description: msg,
		Footer: &discordgo.MessageEmbedFooter{
			Text: doc.URL,
		},
	}
}

// queryGlobResponse is the same as queryResponse but it allows globbing.
func queryGlobResponse(pkg, name string) *discordgo.MessageEmbed {
	r, err := glob.Compile(name)
	if err != nil {
		return errResponse("error parsing glob pattern")
	}
	doc, err := getDoc(pkg)
	if err != nil {
		return errResponse("Error while fetching the page for the package `%s`", pkg)
	}

	var msg string
	for _, fn := range doc.Functions {
		if fn.Type == docs.FnNormal &&
			r.MatchString(fn.Name) {
			msg += fmt.Sprintf("`%s`\n", fn.Signature)
			if len(fn.Comments) > 0 {
				msg += fn.Comments[0]
			} else {
				msg += "*no information available*"
			}
		}
	}

	for _, t := range doc.Types {
		if r.MatchString(t.Name) {
			msg += fmt.Sprintf("```go\n%s\n```\n", t.Signature)
			if len(t.Comments) > 0 {
				msg += t.Comments[0]
			} else {
				msg += "*no information available*"
			}
		}
	}

	if msg == "" {
		return errResponse("No matches found for the pattern `%s` in package `%s`", name, pkg)
	}

	if len(msg) > 2000 {
		msg = fmt.Sprintf("%s...\n\n*note: the message was trimmed to fit the 2k character limit*", msg[:1950])
	}
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Matches for `%s` in package %s", name, pkg),
		URL:         doc.URL,
		Description: msg,
		Footer: &discordgo.MessageEmbedFooter{
			Text: doc.URL,
		},
	}
}
