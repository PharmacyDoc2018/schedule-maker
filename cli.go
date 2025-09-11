package main

import (
	"os"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

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

func (c *config) readlineSetup() *readline.PrefixCompleter {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("select",
			readline.PcItem("pt",
				readline.PcItemDynamic(c.getPatientArgs),
			),
		),
	)

	return completer
}

func (c *config) getPatientArgs(input string) []string {
	var patients []string
	for _, val := range c.patientList {
		patients = append(patients, val.name)
	}

	return patients
}
