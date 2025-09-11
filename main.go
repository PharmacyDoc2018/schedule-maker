package main

import (
	"fmt"

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

	completer := config.readlineSetup()
	rl, _ := readline.NewEx(&readline.Config{
		Prompt:       config.location.Path(),
		AutoComplete: completer,
	})

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		err = config.CommandExe(line)
		if err != nil {
			fmt.Println(err)
		}
		rl.SetPrompt(config.location.Path())
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
