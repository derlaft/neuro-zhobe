package main

import (
	"fmt"
	"glb"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	// all message handlers are there
	// init them in separate file's init()
	msgHandlers []messageHandler

	// all onConfig handlers are there
	configLoadedHandlers []func()

	// all the running toads are stored in there
	toads     = map[string]*NeuroZhobe{}
	toadsSync sync.RWMutex

	// global configuration
	config *NeuroConfig
)

type (
	messageHandler struct {
		priority uint // the more priority is, the more important it is
		cb       func(*NeuroZhobe, *glb.MUCMessage) (bool, error)
	}

	NeuroZhobe struct {
		bot     *glb.GBot
		admins  map[string]bool
		onlines map[string]bool
		config  *Config
	}

	NeuroConfig struct {
		Zhobe     map[string]Config
		GsendHTTP string `yaml:"gsend_http"`
	}

	Config struct {
		Jabber         *glb.Config
		Root           string
		GsendSecret    string        `yaml:"gsend_secret"`
		RestartTimeout time.Duration `yaml:"restart_timeout"`
	}

	PublicError error
)

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

	for _, handler := range msgHandlers {
		match, err := handler.cb(z, msg)
		if err != nil {
			if _, public := err.(PublicError); public {
				// public errors can be directly sent to chat
				z.bot.Send(fmt.Sprintf("%v: %v", msg.From, err.Error()))
			} else {
				// any other error is considered private
				// and sent only to OP to PM
				z.bot.Send(fmt.Sprintf("%v: 542 SHIT HAPPEND", msg.From))
				if z.admins[msg.From] {
					z.bot.SendPrivate(err.Error(), msg.From)
				}
				return
			}
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

func prepareHandlers() {
	// sort slice by its priority
	sort.Slice(msgHandlers, func(i, j int) bool {
		return msgHandlers[i].priority > msgHandlers[j].priority
	})
}

func main() {

	loadedConfig, err := readConfig()
	if err != nil {
		log.Fatal("Could not read config:", err)
	}
	config = loadedConfig
	for _, cb := range configLoadedHandlers {
		go cb()
	}

	// bind shut-down
	var (
		sigs = make(chan os.Signal, 1)
		stop bool
		done sync.WaitGroup
	)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for name, cfg := range config.Zhobe {
		var (
			copy     = cfg
			nameCopy = name
		)

		if copy.RestartTimeout == 0 {
			copy.RestartTimeout = time.Second * 2
		}

		var zhobe = &NeuroZhobe{
			admins:  make(map[string]bool),
			onlines: make(map[string]bool),
			config:  &copy,
		}

		done.Add(1)
		go func() {

			for !stop {

				zhobe.bot = glb.New(zhobe)
				zhobe.bot.Connect(copy.Jabber)

				// store this toad
				toadsSync.Lock()
				toads[nameCopy] = zhobe
				toadsSync.Unlock()

				zhobe.bot.Wait()

				// unstore this toad so gsend won't work until it's really connected
				toadsSync.Lock()
				delete(toads, nameCopy)
				toadsSync.Unlock()

				zhobe.bot.Free()

				// wait before reconnecting
				if !stop {
					time.Sleep(copy.RestartTimeout)
				}
			}

			done.Done()
		}()

	}

	// wait until termination signal
	<-sigs
	stop = true

	// call all toads for sleep
	toadsSync.RLock()
	for _, toad := range toads {
		toad.bot.Disconnect()
	}
	toadsSync.RUnlock()

	// wait until all the toads are shutted down
	done.Wait()

}
