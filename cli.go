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

	config.readlineCompleterMap = map[int]*readline.PrefixCompleter{}

	return config
}

func (c *config) readlineSetup() *readline.Instance {
	completerMode := make(map[int]*readline.PrefixCompleter)

	completerMode[int(Home)] = readline.NewPrefixCompleter(
		readline.PcItem("select",
			readline.PcItem("pt",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("patient",
				readline.PcItemDynamic(c.getPatientArgs),
			),
		),
	)

	completerMode[int(PatientLoc)] = readline.NewPrefixCompleter(
		readline.PcItem("add",
			readline.PcItem("order"),
		),
	)

	c.readlineCompleterMap = completerMode

	rl, _ := readline.NewEx(&readline.Config{
		Prompt:       c.location.Path(),
		AutoComplete: completerMode[int(Home)],
	})
	c.readlineConfig = rl.Config.Clone()

	return rl

}

func (c *config) getPatientArgs(input string) []string {
	var patients []string
	for _, val := range c.patientList {
		patients = append(patients, val.name)
	}

	return patients
}
