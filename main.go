package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	commands := getCommands()
	config := initREPL()

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
