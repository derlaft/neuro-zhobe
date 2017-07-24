package main

import (
	"fmt"
	"glb"
	"time"
)

var startupTime = time.Now()

func init() {
	commands["uptime"] = uptimeCmd
}

func uptimeCmd(z *NeuroZhobe, msg *glb.MUCMessage, params string) error {

	z.bot.Send(fmt.Sprintf("%v: %v", msg.From, time.Since(startupTime)))
	return nil
}
