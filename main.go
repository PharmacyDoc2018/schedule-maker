package main

import (
	"fmt"

	"github.com/chzyer/readline"
)

type config struct {
	missingOrders        missingOrdersQueue
	lastInput            []string
	pathToSch            string
	location             Location
	patientList          map[string]Patient
	commands             commandMapList
	readlineConfig       *readline.Config
	readlineCompleterMap map[int]*readline.PrefixCompleter
}

func main() {
	config := initREPL()

	err := initScheduledPatients(config)
	if err != nil {
		fmt.Println(err)
	}

	rl := config.readlineSetup()

	fmt.Println("first patient with missing order: ", config.patientList[config.missingOrders.NextPatient()].name)

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		err = config.CommandExe(line)
		if err != nil {
			fmt.Println(err)
		}
		config.resetPrefixCompleterMode()
		rl.SetConfig(config.readlineConfig)
		rl.SetPrompt(config.location.Path())
		fmt.Println()
	}

}
