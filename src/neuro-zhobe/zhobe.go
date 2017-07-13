package main

import (
	"fmt"
	"glb"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// all message handlers are there
// init them in separate file's init()
var handlers []messageHandler

type messageHandler func(*NeuroZhobe, *glb.MUCMessage) (bool, error)

type NeuroZhobe struct {
	bot     *glb.GBot
	admins  map[string]bool
	onlines map[string]bool
	config  *NeuroConfig
}

type NeuroConfig struct {
	Jabber *glb.Config
	Zhobe  struct {
		Root     string
		FIFOPath string `yaml:"fifo_path"`
	}
}

func (z *NeuroZhobe) OnConnect() {
	log.Println("Connected to server")
}

func (z *NeuroZhobe) OnDisconnect(err error) {
	log.Printf("Disconnected from server (err=%v)", err)
}

func (z *NeuroZhobe) OnMUCPresence(p *glb.MUCPresence) {
	z.admins[p.Nick] = p.Online && p.Admin
	z.onlines[p.Nick] = p.Online
}

func (z *NeuroZhobe) OnMUCMessage(msg *glb.MUCMessage) {

	if msg.History {
		return // skip old messags
	}

	// Log message first
	log.Printf("%v: %v", msg.From, msg.Body)

	if msg.From == z.bot.Nickname() {
		return // skip self messages
	}

	for _, handler := range handlers {
		match, err := handler(z, msg)
		if err != nil {
			z.bot.Send(fmt.Sprintf("%v: 542 SHIT HAPPEND", msg.From))
			if z.admins[msg.From] {
				z.bot.SendPrivate(err.Error(), msg.From)
			}
			return
		}

		if match {
			return
		}
	}
}

func readConfig() (*NeuroConfig, error) {

	var configFileLocation = "./config.yaml"
	if len(os.Args) >= 2 {
		configFileLocation = os.Args[1]
	}

	bytes, err := ioutil.ReadFile(configFileLocation)
	if err != nil {
		return nil, err
	}

	var result NeuroConfig
	err = yaml.Unmarshal(bytes, &result)

	return &result, err
}

func main() {

	config, err := readConfig()
	if err != nil {
		log.Fatal("Could not read config:", err)
	}

	var zhobe = &NeuroZhobe{
		admins:  make(map[string]bool),
		onlines: make(map[string]bool),
		config:  config,
	}

	zhobe.bot = glb.New(zhobe)
	zhobe.bot.Connect(config.Jabber)
	zhobe.bot.Wait()
	zhobe.bot.Free()
}
