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

type cmdHandler func(z *NeuroZhobe, msg *glb.MUCMessage, params string) error

const commandPrefix = "!"

var (
	commandRegexp = regexp.MustCompile(fmt.Sprintf("^%v[^ ]+", regexp.QuoteMeta(commandPrefix)))
	stripRegexp   = regexp.MustCompile("(`|\\$|\\.\\.)")
	quoteRegexp   = regexp.MustCompile("(\"|')")

	commands = map[string]cmdHandler{}
)

func init() {
	msgHandlers = append(msgHandlers, messageHandler{
		priority: 100,
		cb:       commandHandler,
	})
}

func strip(s string) string {
	return quoteRegexp.ReplaceAllString(stripRegexp.ReplaceAllString(s, ""), "“")
}

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

	handler, builtin := commands[command]
	if builtin && handler != nil {
		return true, handler(z, msg, params)
	}

	search := path.Join(z.config.Root, "./plugins/", path.Base(command))
	// check if file exists
	if _, err := os.Stat(search); os.IsNotExist(err) {
		return true, PublicError(fmt.Errorf("%v: WAT", msg.From))
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

// return values: stdout, true if path binary/script is found, err if any
func (z *NeuroZhobe) executePlugin(path, from, args string, isAdmin bool) (string, error) {

	return z.execute(path, from, fmt.Sprintf("%v", isAdmin), args)
}

func (z *NeuroZhobe) execute(path string, args ...string) (string, error) {
	// exec and get output

	filteredArgs := make([]string, len(args))
	for i, arg := range args {
		filteredArgs[i] = strip(arg)
	}

	cmd := exec.Command(path, filteredArgs...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	out, _ := ioutil.ReadAll(stdout)
	outerr, _ := ioutil.ReadAll(stderr)

	err := cmd.Wait()
	if err != nil {
		return "", err
	}

	err = nil

	if len(outerr) != 0 {
		err = fmt.Errorf("%s", outerr)
	}

	return strings.TrimRight(string(out), " \t\n"), err
}
