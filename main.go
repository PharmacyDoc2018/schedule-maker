package main

import (
	"fmt"

	"github.com/chzyer/readline"
)

type config struct {
	missingOrders        missingOrdersQueue
	lastInput            []string
	IgnoredOrders        IgnoredOrders
	PrepullOrders        PrepullOrders
	pathToSch            string
	pathToSave           string
	pathToIgnoredOrders  string
	pathToPrepullOrders  string
	location             Location
	PatientList          map[string]Patient `json:"patient_list"`
	patientNameMap       map[string]struct{}
	commands             commandMapList
	readlineConfig       *readline.Config
	readlineCompleterMap map[int]*readline.PrefixCompleter
	rl                   *readline.Instance
}

func main() {
	config := initREPL()

	err := initScheduledPatients(config)
	if err != nil {
		fmt.Println(err)
	}

	config.rl = config.readlineSetup()

	for {
		config.readlineLoopStartPreprocess()
		line, err := config.rl.Readline()
		if err != nil {
			break
		}
		err = config.CommandExe(line)
		if err != nil {
			fmt.Println(err)
		}
		config.resetPrefixCompleterMode()
		config.rl.SetConfig(config.readlineConfig)
		config.rl.SetPrompt(config.location.Path())
		fmt.Println()
	}

}
