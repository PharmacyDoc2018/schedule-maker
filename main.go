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
	completer     *readline.PrefixCompleter
}

func main() {
	config := initREPL()

	err := initScheduledPatients(config)
	if err != nil {
		fmt.Println(err)
	}

	config.readlineSetup()
	rl, _ := readline.NewEx(&readline.Config{
		Prompt:       config.location.Path(),
		AutoComplete: config.completer,
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
		rl.SetAutoComplete(config.completer) //-- Just need to have it reset the entir config
		fmt.Println()                        //-- Which is both prompt and completer
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
