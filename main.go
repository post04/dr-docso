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
	c        Config
	commands = []string{"help", "docs"}
)

func ready(session *discordgo.Session, evt *discordgo.Ready) {
	fmt.Printf("Logged in under: %v#%v\n", evt.User.Username, evt.User.Discriminator)
	session.UpdateGameStatus(0, fmt.Sprintf("%vhelp for information!", c.Prefix))
	go cmd.CheckListeners()
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
	cmd.DocsHelpEmbed.Description = fmt.Sprintf("__**Examples:**__ \n%vdocs strings\n%vdocs strings equalsfold\n%vdocs strings builder\n%vdocs strings builder.*\n%vdocs strings *.writestring", c.Prefix, c.Prefix, c.Prefix, c.Prefix, c.Prefix)
	bot, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		log.Fatal("ERROR LOGGING IN", err)
	}
	bot.AddHandler(ready)
	bot.AddHandler(cmd.ReactionListen)

	commandHandlerStruct := New(c.Prefix, true)
	commandHandlerStruct.AddCommand("docs", "{prefix}docs github.com/bwmarrin/discordgo", "Get the documentation of a package from pkg.go.dev", cmd.HandleDoc)
	commandHandlerStruct.AddCommand("getfuncs", "{prefix}getfuncs github.com/bwmarrin/discordgo", "Get all the functions in a package from pkg.go.dev", cmd.FuncsPages)
	commandHandlerStruct.AddCommand("gettypes", "{prefix}gettypes github.com/bwmarrin/discordgo", "Get all the types in a package from pkg.go.dev", cmd.TypesPages)
	commandHandlerStruct.GenHelp()
	bot.AddHandler(commandHandlerStruct.OnMessage)
	err = bot.Open()
	if err != nil {
		log.Fatal("ERROR OPENING CONNECTION", err)
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	bot.Close()
}
