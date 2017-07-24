package main

import (
	"fmt"
	"glb"
)

func init() {
	commands["megakick"] = megakickCmd
}

func megakickCmd(z *NeuroZhobe, msg *glb.MUCMessage, who string) error {

	// you can't kick a cockroach
	if who <= "" {
		return PublicError(fmt.Errorf("WAT"))
	}

	// there we make various checks, but in general we have no way to find out if kick fails for now
	if z.admins[who] || !z.onlines[who] || !z.admins[z.bot.Nickname()] {
		return PublicError(fmt.Errorf("Can't megakick %v", who))
	}

	// only admin is able to kick
	if !z.admins[msg.From] {
		return PublicError(fmt.Errorf("GTFO"))
	}

	z.bot.Kick(who, "megakick")
	return nil
}
