package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/thatguydoru/nootify"
)

var (
	TOKEN          string
	TARGET_MESSAGE string
	TARGET_CHANNEL string
	TARGET_ROLE    string
)

func init() {
	flag.StringVar(&TOKEN, "--token", "", "the bot token")
	flag.StringVar(&TOKEN, "--message", "", "the target message")
	flag.StringVar(&TOKEN, "--channel", "", "the target channel")
	flag.StringVar(&TOKEN, "--role", "", "the target role")
    flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatalln("failed to create a discord session:", err)
	}
	dg.Identify.Intents |= discordgo.IntentGuilds
	dg.Identify.Intents |= discordgo.IntentsGuildMessages
	dg.Identify.Intents |= discordgo.IntentsGuildMembers

	message, err := dg.ChannelMessage(TARGET_CHANNEL, TARGET_MESSAGE)
	if err != nil {
		log.Fatalln("failed to get message:", err)
	}

    // The Nootify begins
	notifier := nootify.InitNootify(dg, message, '!')
	err = notifier.RegisterNootOption(
		"😔",
		TARGET_ROLE,
		"man",
		"go watch https://twitch.com/unitoftime",
	)
	if err != nil {
		log.Fatalln("noot not registered:", err)
	}
	notifier.GoNoot()
    // The Nootify ends

	if err := dg.Open(); err != nil {
		log.Fatalln("failed to open connection:", err)
	}
	defer dg.Close()

	fmt.Println("Bot is now running. To exit press Ctrl + C")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
