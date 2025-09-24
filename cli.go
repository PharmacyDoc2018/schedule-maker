package main

import (
	"os"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

func initREPL() *config {
	godotenv.Load(".env")
	pathToSch := os.Getenv("SCH_PATH")
	pathToSave := os.Getenv("SAVE_PATH")

	config := &config{
		pathToSch:  pathToSch,
		pathToSave: pathToSave,
	}

	config.commands = getCommands()

	config.PatientList = map[string]Patient{}

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
		readline.PcItem("exit"),
		readline.PcItem("get",
			readline.PcItem("schedule",
				readline.PcItem("infusion"),
				readline.PcItem("-i"),
			),
			readline.PcItem("next",
				readline.PcItem("mop"),
				readline.PcItem("missingOrderPatient"),
			),
			readline.PcItem("clear"),
		),
	)

	completerMode[int(PatientLoc)] = readline.NewPrefixCompleter(
		readline.PcItem("add",
			readline.PcItem("order"),
		),
		readline.PcItem("home"),
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
	for _, val := range c.PatientList {
		patients = append(patients, val.Name)
	}

	return patients
}

func (c *config) resetPrefixCompleterMode() {
	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		c.readlineConfig.AutoComplete = c.readlineCompleterMap[int(Home)]

	case PatientLoc:
		c.readlineConfig.AutoComplete = c.readlineCompleterMap[int(PatientLoc)]
	}
}
