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

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		err = config.CommandExe(line)
		if err != nil {
			fmt.Println(err)
		}
		//rl.SetPrompt(config.location.Path())
		rl.SetConfig(config.readlineConfig)
		fmt.Println()
	}

	/**
	scheduleMaker := bufio.NewScanner(os.Stdin)

	fmt.Printf(config.location.Path())
	//fmt.Printf("pharmacy > ")
	for scheduleMaker.Scan() {
		input := scheduleMaker.Text()
		err = config.CommandExe(input)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
		fmt.Printf(config.location.Path())
		//fmt.Printf("pharmacy > ")
	}
		**/

}
