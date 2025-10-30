package main

import (
	"fmt"
	"time"

	"github.com/chzyer/readline"
)

type config struct {
	missingOrders        missingOrdersQueue
	lastInput            []string
	PrepullOrders        PrepullOrders
	lastSave             time.Time
	pathToSch            string
	pathToSave           string
	pathToIgnoredOrders  string
	pathToPrepullOrders  string
	pathToPtSupplyOrders string
	location             Location
	IgnoredOrders        IgnoredOrders
	PtSupplyOrders       PtSupplyOrders
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

	config.createPatientNameMap()

	config.FindMissingInfusionOrders()

	err = config.PullIgnoredOrdersList()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = config.PullPrepullOrdersList()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = config.PullPtSupplyOrdersList()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println()

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
