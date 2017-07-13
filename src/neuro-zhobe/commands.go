package main

/*
	Command handler. Takes care of stuff like !help and !megakick
*/

import (
	"fmt"
	"glb"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

func init() {
	handlers = append(handlers, commandHandler)
}

const commandPrefix = "!"

var commandRegexp = regexp.MustCompile(fmt.Sprintf("^%v[^ ]+", regexp.QuoteMeta(commandPrefix)))

func commandHandler(z *NeuroZhobe, msg *glb.MUCMessage) (bool, error) {

	// check for command regexp
	if !commandRegexp.MatchString(msg.Body) {
		return false, nil
	}

	// split command message into parts
	var (
		tokens  = strings.SplitN(msg.Body, " ", 2)
		command string
		params  string
	)

	if len(tokens) > 0 && len(tokens[0]) > 0 {
		command = tokens[0][1:] // command without a prefix (!>help< coco)
	}

	if len(tokens) > 1 {
		params = tokens[1] // params  (!help >cococo coco co<)
	}

	switch command {
	case "megakick": // implement kick as a built-in

		// there we make various checks, but in general we have no way to find out if kick fails for now
		if z.admins[params] || !z.onlines[params] || !z.admins[z.bot.Nickname()] {
			return true, fmt.Errorf("Can't megakick %v", params)
		} else if !z.admins[msg.From] {
			return true, fmt.Errorf("GTFO")
		}

		z.bot.Kick(params, "megakick")
		return true, nil

	default:

		search := path.Join(z.config.Zhobe.Root, "./plugins/", path.Base(command))
		// check if file exists
		if _, err := os.Stat(search); os.IsNotExist(err) {
			return true, fmt.Errorf("WAT")
		}

		// execute plugin file
		result, err := z.executePlugin(search, msg.From, params, z.admins[msg.From])
		if result > "" {
			z.bot.Send(result)
		}
		if err != nil {
			return true, err
		}

		return true, nil
	}

	return true, fmt.Errorf("WAT")
}

// return values: stdout, true if path binary/script is found, err if any
func (z *NeuroZhobe) executePlugin(path, from, args string, isAdmin bool) (string, error) {

	return z.execute(path, from, fmt.Sprintf("%v", isAdmin), args)
}

func (z *NeuroZhobe) execute(path string, args ...string) (string, error) {
	// exec and get output

	cmd := exec.Command(path, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	out, _ := ioutil.ReadAll(stdout)
	outerr, _ := ioutil.ReadAll(stderr)

	cmd.Wait()

	var err error = nil

	if len(outerr) != 0 {
		err = fmt.Errorf("%s", outerr)
	}

	return strings.TrimRight(string(out), " \t\n"), err
}
