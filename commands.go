package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type commandMapList map[string]cliCommand

func getCommands() commandMapList {
	commands := commandMapList{
		"hello": {
			name:        "hello",
			description: "prints hello to the console",
			callback:    commandHello,
		},
		"select": {
			name:        "select",
			description: "select a location to move to",
			callback:    commandSelect,
		},
	}
	return commands
}

func initREPL() *config {
	godotenv.Load(".env")
	pathToSch := os.Getenv("SCH_PATH")

	config := &config{
		pathToSch: pathToSch,
	}

	config.commands = getCommands()

	config.patientList = map[string]Patient{}

	nodeMap := make(map[int]*LocationNode)
	nodeMap[0] = &LocationNode{
		id:       0,
		name:     "pharmacy",
		locType:  Home,
		parentID: -1,
	}
	config.location = Location{
		allNodes:      nodeMap,
		currentNodeID: 0,
	}

	return config
}

func initPrefixCompleter() *readline.PrefixCompleter {
	completer := readline.NewPrefixCompleter()
	fmt.Println(completer.Name)
	fmt.Println(completer.Dynamic)
	fmt.Println(completer.Callback)
	fmt.Println(completer.Children)

	return completer

}

func cleanInput(text string) []string {
	var textWords []string
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	firstPass := strings.Split(text, " ")

	for _, word := range firstPass {
		if word != "" {
			textWords = append(textWords, word)
		}
	}
	return textWords
}

func cleanInputAndStore(c *config, input string) {
	c.lastInput = cleanInput(input)
}

func commandHello(c *config) error {
	fmt.Println("Hello, World!")
	return nil
}

func commandSelect(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error too few arguments")
	}

	if len(c.lastInput) > 3 {
		return fmt.Errorf("error too many arguments")
	}

	firstArg := c.lastInput[1]

	switch firstArg {
	case "pt", "patient":
		err := commandSelectPatient(c)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func commandSelectPatient(c *config) error {
	mrn := c.lastInput[2]
	if _, ok := c.patientList[mrn]; !ok {
		return fmt.Errorf("error. cannot find patient")
	}
	err := c.location.SelectPatientNode(mrn)
	if err != nil {
		return err
	}

	return nil
}

func (c *config) commandLookup(input string) (cliCommand, error) {
	for _, c := range c.commands {
		if input == c.name {
			return c, nil
		}
	}
	return cliCommand{}, fmt.Errorf("unknown command")
}

func (c *config) CommandExe(input string) error {
	cleanInputAndStore(c, input)
	command, err := c.commandLookup(c.lastInput[0])
	if err != nil {
		return err
	}
	command.callback(c)
	return nil
}
