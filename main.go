package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/chzyer/readline"
)

type config struct {
	missingOrders missingOrdersQueue
	lastInput     []string
	pathToSch     string
	location      Location
	patientList   map[string]Patient
	commands      commandMapList
}

func main() {
	config := initREPL()

	err := initScheduledPatients(config)
	if err != nil {
		fmt.Println(err)
	}

	scheduleMaker := bufio.NewScanner(os.Stdin)

	completer := initPrefixCompleter()
	_, err = readline.NewEx(&readline.Config{
		AutoComplete: completer,
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf(config.location.Path())
	//fmt.Printf("pharmacy > ")
	for scheduleMaker.Scan() {
		input := scheduleMaker.Text()
		err := config.CommandExe(input)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
		fmt.Printf(config.location.Path())
		//fmt.Printf("pharmacy > ")
	}

}
