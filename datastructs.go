package main

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	scheduleRows [][]string
	ordersRows   [][]string
	lastInput    []string
	pathToSch    string
}

type commandMapList map[string]cliCommand
