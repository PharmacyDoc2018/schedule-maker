package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	commands := getCommands()
	config := initREPL()

	err := pullData(config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config.scheduleRows[0][0])
	fmt.Println(config.scheduleRows[1][0])
	fmt.Println(config.scheduleRows[2][0])

	fmt.Println(config.ordersRows[0][0])
	fmt.Println(config.ordersRows[1][0])
	fmt.Println(config.ordersRows[2][0])

	scheduleMaker := bufio.NewScanner(os.Stdin)
	fmt.Printf("pharmacy > ")
	for scheduleMaker.Scan() {
		input := scheduleMaker.Text()
		err := commandExe(input, commands, config)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
		fmt.Printf("pharmacy > ")
	}

}
