package glb

// #cgo LDFLAGS: -lstdc++ -lgloox
// #include "gloox.h"
import "C"

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	registry     = map[C.GBot]*GBot{}
	registryLock sync.RWMutex
)

type (
	GBot struct {
		sync.Mutex

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
		Self   bool
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
func goSched(cobj C.GBot) {
	bot := instance(cobj)

	bot.Unlock() // give some time to do stuff
	time.Sleep(time.Millisecond * 100)
	bot.Lock()
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
				go bot.PingRoom()

				if time.Since(bot.lastPong) > bot.config.IQTimeout {
					fmt.Println("disconnectin")
					bot.Disconnect()
				}
			}
		})

	}()
}

func (b *GBot) PingRoom() {
	b.Lock()
	defer b.Unlock()

	C.BotPingRoom(b.cobj)
}

//export goOnDisconnect
func goOnDisconnect(cobj C.GBot, errCode, authErr C.int) {
	bot := instance(cobj)

	var err error
	if errCode > 0 || authErr > 0 {
		err = DisconnectError{
			ConnectionError:     ConnectionError(errCode),
			AuthenticationError: AuthenticationError(authErr),
		}
	}

	if cb, ok := bot.cb.(OnDisconnect); ok {
		cb.OnDisconnect(err)
	}

	bot.done <- true
}

//export goOnMessage
func goOnMessage(cobj C.GBot, raw_from, raw_msg *C.char, raw_history, raw_private bool) {

	var (
		bot     = instance(cobj)
		from    = C.GoString(raw_from)
		msg     = C.GoString(raw_msg)
		history = raw_history
		private = raw_private
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
func goOnPresence(cobj C.GBot, raw_nick *C.char, raw_self, raw_presence, raw_affiliation, raw_role C.int) {

	var (
		bot         = instance(cobj)
		nick        = C.GoString(raw_nick)
		self        = raw_self > 0
		presence    = PresenceType(raw_presence)
		affiliation = Affiliation(raw_affiliation)
		role        = Role(raw_role)
	)

	go func() {

		var (
			online = false
			admin  = false
		)

		switch presence {
		case PresenceAvailable, PresenceChat,
			PresenceAway, PresenceDND, PresenceXA:

			online = true
		}

		admin = role == RoleModerator &&
			(affiliation == AffiliationOwner || affiliation == AffiliationAdmin)

		// we must do something when conference is exp. weirdo issues.
		// let's just disconnect for now
		if self {
			switch presence {
			case PresenceInvalid, PresenceError, PresenceUnavailable:
				go bot.Disconnect()
			}
		}

		if cb, ok := bot.cb.(OnMUCPresence); ok {
			cb.OnMUCPresence(&MUCPresence{
				Nick:   nick,
				Online: online,
				Admin:  admin,
				Self:   self,
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
	b.Lock()
	defer b.Unlock()

	C.BotFree(b.cobj)
}

func (b *GBot) Connect(config *Config) {
	b.config = config

	if b.config.IQTimeout == 0 {
		b.config.IQTimeout = time.Second * 10
	}

	go func() {
		b.Lock()
		C.BotConnect(
			b.cobj,
			C.CString(config.JID),
			C.CString(config.Password),
			C.CString(fmt.Sprintf("%v/%v", config.Conference, config.Nickname)),
		)
		log.Println("terminated")
		// wait for termination
		b.Unlock()
	}()
}

func (b *GBot) Disconnect() {
	b.Lock()
	defer b.Unlock()

	if !b.disconnecting {
		b.disconnecting = true
		C.BotDisconnect(b.cobj)
	}
}

func (b *GBot) Nickname() string {
	b.Lock()
	nick := C.BotNick(b.cobj)
	b.Unlock()

	return C.GoString(nick)
}

func (b *GBot) Send(message string) {

	b.Lock()
	defer b.Unlock()

	C.BotReply(
		b.cobj,
		C.CString(message),
	)
}

func (b *GBot) SendPrivate(message, recipient string) {
	b.Lock()
	defer b.Unlock()

	C.BotReplyPrivate(
		b.cobj,
		C.CString(message),
		C.CString(recipient),
	)
}

func (b *GBot) Kick(who, forWhat string) {
	b.Lock()
	defer b.Unlock()

	C.BotKick(
		b.cobj,
		C.CString(who),
		C.CString(forWhat),
	)
}

func (b *GBot) Wait() {
	<-b.done
}
