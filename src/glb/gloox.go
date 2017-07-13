package glb

// #cgo LDFLAGS: -lstdc++ -lgloox
// #include "gloox.h"
import "C"

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

var (
	registry     = map[C.GBot]*GBot{}
	registryLock sync.RWMutex
)

type (
	GBot struct {
		cobj          C.GBot
		config        *Config
		done          chan bool
		cb            interface{}
		lastPong      time.Time
		healthCheck   sync.Once
		disconnecting bool
	}

	Config struct {
		JID        string
		Password   string
		Conference string
		Nickname   string
		SkipTLS    bool          `yaml:"skip_tls"` // all hail to cx
		IQTimeout  time.Duration `yaml:"iq_timeout"`
	}

	MUCMessage struct {
		Body    string
		From    string
		History bool
		Private bool
	}

	MUCPresence struct {
		Nick   string
		Online bool
		Admin  bool
	}

	// callback interfaces
	OnConnect interface {
		OnConnect()
	}

	OnDisconnect interface {
		OnDisconnect(error)
	}

	OnMUCMessage interface {
		OnMUCMessage(*MUCMessage)
	}

	OnMUCPresence interface {
		OnMUCPresence(*MUCPresence)
	}

	OnMUCSubject interface {
		OnMUCSubject(from, subject string)
	}
)

// Get go bot object reference by C void pointer
// this is needed because of shit cgo pointer rules
func instance(cobj C.GBot) *GBot {
	registryLock.RLock()
	defer registryLock.RUnlock()
	ret, ok := registry[cobj]
	if !ok {
		panic("gloox: trying to callback unexisting object")
	}
	return ret
}

func New(cb interface{}) *GBot {
	ret := &GBot{
		done: make(chan bool, 1),
		cobj: C.BotInit(),
		cb:   cb,
	}
	registryLock.Lock()
	registry[ret.cobj] = ret
	registryLock.Unlock()

	return ret
}

// Callbacks

//export goOnTLSConnect
func goOnTLSConnect(cobj C.GBot, status C.int) C.int {
	bot := instance(cobj)

	if status == 0 || bot.config.SkipTLS == true {
		return C.int(1)
	}

	return C.int(0)
}

//export goSched
func goSched() {
	runtime.Gosched()
}

//export goOnConnect
func goOnConnect(cobj C.GBot) {
	bot := instance(cobj)
	go func() {

		if cb, ok := bot.cb.(OnConnect); ok {
			go cb.OnConnect()
		}

		go bot.healthCheck.Do(func() {
			// allow some startup
			bot.lastPong = time.Now()

			for range time.Tick(time.Second) {
				go C.BotPingRoom(bot.cobj)

				if time.Since(bot.lastPong) > bot.config.IQTimeout {
					fmt.Println("disconnectin")
					bot.Disconnect()
				}
			}
		})

	}()
}

//export goOnDisconnect
func goOnDisconnect(cobj C.GBot, errCode, authErr C.int) {
	bot := instance(cobj)

	var err error
	if errCode > 0 || authErr > 0 {
		err = fmt.Errorf("gld: dissonnected with error (errCode=%v, authError=%v)", errCode, authErr)
	}

	go func() {

		if cb, ok := bot.cb.(OnDisconnect); ok {
			fmt.Println("ehoh", bot.disconnecting)

			cb.OnDisconnect(err)
		}

		bot.done <- true
	}()
}

//export goOnMessage
func goOnMessage(cobj C.GBot, raw_from, raw_msg *C.char, history, private bool) {

	var (
		bot  = instance(cobj)
		from = C.GoString(raw_from)
		msg  = C.GoString(raw_msg)
	)

	go func() {
		if cb, ok := bot.cb.(OnMUCMessage); ok {
			cb.OnMUCMessage(&MUCMessage{
				Body:    msg,
				From:    from,
				History: history,
				Private: private,
			})
		}
	}()
}

//export goOnPresence
func goOnPresence(cobj C.GBot, raw_nick *C.char, raw_online, raw_admin bool) {

	var (
		bot    = instance(cobj)
		nick   = C.GoString(raw_nick)
		online = raw_online
		admin  = raw_admin
	)

	go func() {

		if cb, ok := bot.cb.(OnMUCPresence); ok {
			cb.OnMUCPresence(&MUCPresence{
				Nick:   nick,
				Online: online,
				Admin:  admin,
			})
		}
	}()
}

//export goOnMUCSubject
func goOnMUCSubject(cobj C.GBot, raw_nick, raw_subject *C.char) {

	var (
		bot     = instance(cobj)
		nick    = C.GoString(raw_nick)
		subject = C.GoString(raw_subject)
	)

	go func() {
		if cb, ok := bot.cb.(OnMUCSubject); ok {
			cb.OnMUCSubject(nick, subject)
		}
	}()
}

//export goOnPing
func goOnPing(cobj C.GBot, success bool) {
	var bot = instance(cobj)

	if success {
		bot.lastPong = time.Now()
	}
}

//export goOnError
func goOnError(cobj C.GBot, errcode int) {

	var bot = instance(cobj)
	_ = bot
	log.Println("Got muc error: %v", errcode)
}

func (b *GBot) Free() {
	C.BotFree(b.cobj)
}

func (b *GBot) Connect(config *Config) {
	b.config = config

	if b.config.IQTimeout == 0 {
		b.config.IQTimeout = time.Second * 10
	}

	go func() {

		C.BotConnect(
			b.cobj,
			C.CString(config.JID),
			C.CString(config.Password),
			C.CString(fmt.Sprintf("%v/%v", config.Conference, config.Nickname)),
		)
	}()
}

func (b *GBot) Disconnect() {
	b.disconnecting = true
	C.BotDisconnect(b.cobj)
}

func (b *GBot) Nickname() string {
	return C.GoString(C.BotNick(b.cobj))
}

func (b *GBot) Send(message string) {
	C.BotReply(
		b.cobj,
		C.CString(message),
	)
}

func (b *GBot) SendPrivate(message, recipient string) {
	C.BotReplyPrivate(
		b.cobj,
		C.CString(message),
		C.CString(recipient),
	)
}

func (b *GBot) Kick(who, forWhat string) {
	C.BotKick(
		b.cobj,
		C.CString(who),
		C.CString(forWhat),
	)
}

func (b *GBot) Wait() {
	<-b.done
}
