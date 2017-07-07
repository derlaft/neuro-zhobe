package main

import (
	"fmt"
	"glb"
	"regexp"
)

func init() {
	handlers = append(handlers, callHandler)
}

func (z *NeuroZhobe) CallRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s[:,][ \t]*", regexp.QuoteMeta(z.bot.Nickname())))
}

func callHandler(z *NeuroZhobe, msg *glb.MUCMessage) (bool, error) {
	if found := z.CallRegexp().FindStringIndex(msg.Body); len(found) >= 2 {
		messageBody := msg.Body[found[1]:]
		answer, err := z.execute("./chat/answer", msg.From, messageBody)
		if err != nil {
			return true, err
		}
		z.bot.Send(answer)
		return true, nil
	}

	return false, nil
}
