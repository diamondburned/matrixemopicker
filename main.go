package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/chanbakjsd/gotrix"
	"github.com/chanbakjsd/gotrix/event"
	"github.com/diamondburned/matrixemopicker/app/components/login"
	"github.com/gotk3/gotk3/gtk"
)

type EmotesEvent map[string]interface{}

var _ event.Event = (*EmotesEvent)(nil)

func (ev EmotesEvent) Type() event.Type { return "im.ponies.user_emotes" }

var (
	sigint = make(chan os.Signal)
	bgWait sync.WaitGroup
)

func main() {
	signal.Notify(sigint, os.Interrupt)

	gtk.Init(&os.Args)

	login := login.NewLogin(onClient)
	login.Show()

	gtk.Main()
	bgWait.Wait()
}

func onClient(client *gotrix.Client) {
	bgWait.Add(1)
	go func() {
		open(client)
		bgWait.Done()
	}()
	gtk.MainQuit()
}

func open(client *gotrix.Client) {
	client.AddHandler(func(client *gotrix.Client, ev EmotesEvent) {
		log.Printf("%#v", ev)
	})

	if err := client.Open(); err != nil {
		log.Fatalln("failed to open:", err)
	}

	log.Println("Opened client.")

	defer client.Close()

	<-sigint
}
