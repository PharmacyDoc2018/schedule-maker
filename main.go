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
