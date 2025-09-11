package main

import (
	"fmt"
	"strings"
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
