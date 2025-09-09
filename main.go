package main

import (
	"bufio"
	"fmt"
	"os"
)

type config struct {
	missingOrders missingOrdersQueue
	lastInput     []string
	pathToSch     string
	patientList   map[string]Patient
	location      Location
}

func main() {
	commands := getCommands()
	config := initREPL()

	err := initScheduledPatients(config)
	if err != nil {
		fmt.Println(err)
	}

	scheduleMaker := bufio.NewScanner(os.Stdin)
	fmt.Printf(config.location.Path())
	//fmt.Printf("pharmacy > ")
	for scheduleMaker.Scan() {
		input := scheduleMaker.Text()
		err := commandExe(input, commands, config)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
		fmt.Printf(config.location.Path())
		//fmt.Printf("pharmacy > ")
	}

}
