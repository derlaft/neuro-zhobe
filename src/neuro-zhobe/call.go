package main

import (
	"fmt"
	"glb"
	"log"
	"regexp"
)

func init() {
	msgHandlers = append(msgHandlers, messageHandler{
		cb:       callHandler,
		priority: 100,
	})
}

func (z *NeuroZhobe) CallRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s[:,][ \t]*", regexp.QuoteMeta(z.bot.Nickname())))
}

func callHandler(z *NeuroZhobe, msg *glb.MUCMessage) (bool, error) {
	if found := z.CallRegexp().FindStringIndex(msg.Body); len(found) >= 2 {
		var (
			messageBody = msg.Body[found[1]:]
			isAdmin     = fmt.Sprintf("%v", z.admins[msg.From])
		)

		answer, err := z.execute("./chat/answer", msg.From, isAdmin, messageBody)
		if err != nil {
			return true, err
		}

		z.bot.Send(answer)
		return true, nil
	}

	log.Println("coc")
	return false, nil
}
