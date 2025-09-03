package main

import (
	"fmt"
	"strings"
)

func getCommands() commandMapList {
	commands := commandMapList{
		"hello": {
			name:        "hello",
			description: "prints hello to the console",
			callback:    commandHello,
		},
	}
	return commands
}

func initREPL() *config {
	config := &config{}
	return config
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

func commandLookup(input string, commands commandMapList) (cliCommand, error) {
	for _, c := range commands {
		if input == c.name {
			return c, nil
		}
	}
	return cliCommand{}, fmt.Errorf("unknown command")
}

func commandExe(input string, commands commandMapList, config *config) error {
	cleanInputAndStore(config, input)
	command, err := commandLookup(config.lastInput[0], commands)
	if err != nil {
		return err
	}
	command.callback(config)
	return nil
}
