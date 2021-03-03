package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/postrequest69/dr-docso/docs"
)

func HandleDoc(s *discordgo.Session, m *discordgo.MessageCreate) {
	var msg *discordgo.MessageEmbed
	fields := strings.Fields(m.Content)
	switch len(fields) {
	case 0: // probably impossible
		return
	case 1: // only the invocation
		msg = helpShortResponse()
	case 2: // invocation + arg
		msg = pkgResponse(fields[1])
	case 3: // invocation + pkg + func
		msg = funcResponse(fields[1], fields[2])
	case 4:
		switch strings.ToLower(fields[2]) {
		case "func", "function", "fn":
			msg = funcResponse(fields[1], fields[3])
		case "type":
			msg = typeResponse(fields[1], fields[3])
		default:
			msg = errResponse("Unsupported search type %q\nValid options are:\n\t`func`\n\t`type`", fields[2])
		}
	default:
		msg = errResponse("Too many arguments.")
	}

	if msg == nil {
		msg = errResponse("No results found, possibly an internal error.")
	}
	s.ChannelMessageSendEmbed(m.ChannelID, msg)
}

func funcResponse(pkg, name string) *discordgo.MessageEmbed {
	// TODO: implement caching for stdlib
	var err error
	// check if pkg is in stdlib and is cached
	doc, ok := StdlibCache[pkg]
	if !ok || doc == nil {
		doc, err = docs.GetDoc(pkg)
		if err != nil {
			return errResponse("An error occurred while fetching the page for pkg `%s`", pkg)
		}
	}

	// cache the sdlib pkg
	if ok && doc != nil {
		StdlibCache[pkg] = doc
	}
	// cache if it's a stdlib pkg
	if d, ok := StdlibCache[pkg]; ok && d == nil && doc != nil {
		StdlibCache[pkg] = doc
	}

	if len(doc.Functions) == 0 {
		return errResponse("No results found for package: %q, function: %q", pkg, name)
	}

	var msg string

	// TODO(note): maybe use levenshtein here?
	for _, fn := range doc.Functions {
		if strings.EqualFold(fn.Name, name) {
			// match found
			msg += fmt.Sprintf("`%s`", fn.Signature)
			if len(fn.Comments) > 0 {
				msg += fmt.Sprintf("\n%s", fn.Comments[0])
			} else {
				msg += "\n*no information*"
			}
			if fn.Example != "" {
				msg += fmt.Sprintf("\n\nExample:\n```go\n%s\n```", fn.Example)
			}
		}
	}

	if msg == "" {
		return errResponse("The package `%s` does not have function `%s`", pkg, name)
	}
	if len(msg) > 2000 {
		msg = fmt.Sprintf("%s\n\n*note: the message was trimmed to fit the 2k character limit*", msg[:1950])
	}
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s: func %s", pkg, name),
		Description: msg,
	}
}

func errResponse(format string, args ...interface{}) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: fmt.Sprintf(format, args...),
	}
}

func typeResponse(pkg, name string) *discordgo.MessageEmbed {
	doc, err := docs.GetDoc(pkg)
	if err != nil {
		return errResponse("An error occurred while getting the page for the package `%s`", pkg)
	}
	if len(doc.Types) == 0 {
		return errResponse("Package `%s` seems to have no type definitions", pkg)
	}

	var msg string

	for _, t := range doc.Types {
		if strings.EqualFold(t.Name, name) {
			// got a match
			msg += fmt.Sprintf("```go\n%s\n```", t.Signature)
			if len(t.Comments) > 0 {
				msg += fmt.Sprintf("\n%s", t.Comments[0])
			} else {
				msg += "\n*no information*"
			}
		}
	}

	if msg == "" {
		return errResponse("Package `%s` does not have type `%s`", pkg, name)
	}
	if len(msg) > 2000 {
		msg = fmt.Sprintf("%s\n\n*note: the message is trimmed to fit the 2k character limit*", msg[:1950])
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s: type %s", pkg, name),
		Description: msg,
	}
}

func helpShortResponse() *discordgo.MessageEmbed {
	// TODO: implement this
	return nil
}

func pkgResponse(pkg string) *discordgo.MessageEmbed {
	// doc, err
	_, err := docs.GetDoc(pkg)
	if err != nil {
		return errResponse("An error occured when requesting the page for the package `%s`", pkg)
	}
	// TODO: implement this
	return nil
}

func methodResponse(pkg, t, name string) *discordgo.MessageEmbed {
	doc, err := docs.GetDoc(pkg)
	if err != nil {
		return errResponse("Error while getting the page for the package `%s`", pkg)
	}
	if len(doc.Functions) == 0 {
		return errResponse("Package `%s` seems to have no functions", pkg)
	}

	var msg string

	for _, fn := range doc.Functions {
		if fn.Type == docs.FnMethod &&
			strings.EqualFold(fn.Name, name) &&
			strings.EqualFold(fn.MethodOf, t) {
			msg += fmt.Sprintf("`%s`", fn.Signature)
			if len(fn.Comments) > 0 {
				msg += fmt.Sprintf("\n%s", fn.Comments[0])
			} else {
				msg += "\n*no info*"
			}
			if fn.Example != "" {
				msg += fmt.Sprintf("\nExample:\n```\n%s\n```", fn.Example)
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
		Description: msg,
	}
}
