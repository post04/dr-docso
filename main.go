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
	cmd "github.com/post04/dr-docso/bot"
)

var (
	c Config
)

func ready(session *discordgo.Session, evt *discordgo.Ready) {
	fmt.Printf("Logged in under: %v#%v\n", evt.User.Username, evt.User.Discriminator)
	session.UpdateGameStatus(0, fmt.Sprintf("%vhelp for information!", c.Prefix))
	go cmd.CheckListeners()
}

func getConfig() {
	fileBytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(fileBytes, &c)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	getConfig()
	cmd.DocsHelpEmbed.Description = fmt.Sprintf(`__**Examples:**__
%sdocs strings
%sdocs strings equalfold
%sdocs strings builder
%sdocs strings builder.*
%sdocs strings *.writestring`,
		c.Prefix, c.Prefix, c.Prefix, c.Prefix, c.Prefix)
	bot, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		log.Fatal("ERROR LOGGING IN", err)
	}
	bot.AddHandler(ready)
	bot.AddHandler(cmd.ReactionListen)

	cmdhandler := New(c.Prefix, true)
	cmdhandler.AddCommand("docs", "{prefix}docs github.com/bwmarrin/discordgo", "Get the documentation of a package from pkg.go.dev", cmd.HandleDoc)
	cmdhandler.AddCommand("funcs", "{prefix}funcs github.com/bwmarrin/discordgo", "Get all the functions in a package from pkg.go.dev", cmd.HandleFuncsPages)
	cmdhandler.AddCommand("types", "{prefix}types github.com/bwmarrin/discordgo", "Get all the types in a package from pkg.go.dev", cmd.HandleTypesPages)
	cmdhandler.GenHelp()
	bot.AddHandler(cmdhandler.OnMessage)
	err = bot.Open()
	if err != nil {
		log.Fatal("ERROR OPENING CONNECTION", err)
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc
	bot.Close()
}
