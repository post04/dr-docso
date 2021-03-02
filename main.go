package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	c        Config
	commands = []string{"help", "docs"}
)

func ready(session *discordgo.Session, evt *discordgo.Ready) {
	fmt.Printf("Logged in under: %v:%v\n", evt.User.Username, evt.User.Discriminator)
	session.UpdateGameStatus(0, fmt.Sprintf("%vhelp for information!", c.Prefix))
	go checkListeners()
}

func getConfig() {
	fileBytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(fileBytes, &c)
	if err != nil {
		panic(err)
	}
}

func main() {
	getConfig()
	docsCommandHelpEmbed.Description = fmt.Sprintf("__**Examples:**__ \n%vdocs strings\n%vdocs strings equalsfold\n%vdocs strings types builder\n%vdocs strings functions equalsfold", c.Prefix, c.Prefix, c.Prefix, c.Prefix)
	bot, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		log.Fatal("ERROR LOGGING IN", err)
		return
	}
	bot.AddHandler(ready)
	bot.AddHandler(reactionListen)
	commandHandlerStruct := New(c.Prefix, true)
	commandHandlerStruct.AddCommand("docs", "{prefix}docs github.com/bwmarrin/discordgo", "Get the documentation of a package from pkg.go.dev!", DocsCommand)
	bot.AddHandler(commandHandlerStruct.OnMessage)
	err = bot.Open()
	if err != nil {
		log.Fatal("ERROR OPENING CONNECTION", err)
		return
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = bot.Close()
}
