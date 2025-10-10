package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type commandMapList map[string]cliCommand

func getCommands() commandMapList {
	commands := commandMapList{
		"hello": {
			name:        "hello",
			description: "prints hello to the console",
			callback:    commandHello,
		},
		"home": {
			name:        "home",
			description: "returns to home location",
			callback:    commandHome,
		},
		"select": {
			name:        "select",
			description: "select a location to move to",
			callback:    commandSelect,
		},
		"add": {
			name:        "add",
			description: "adds elements depending on location",
			callback:    commandAdd,
		},
		"exit": {
			name:        "exit",
			description: "exists the CLI",
			callback:    commandExit,
		},
		"get": {
			name:        "get",
			description: "get a stored element including schedule and orders",
			callback:    commandGet,
		},
		"clear": {
			name:        "clear",
			description: "clears the screen",
			callback:    commandClear,
		},
		"review": {
			name:        "review",
			description: "opens review node for queues and lists",
			callback:    commandReview,
		},
		"mark": {
			name:        "mark",
			description: "changes a mark for certain items. i.e. mark complete [patient]",
			callback:    commandMark,
		},
		"remove": {
			name:        "remove",
			description: "removes elements depending on location",
			callback:    commandRemove,
		},
		"restart": {
			name:        "restart",
			description: "restarts the program",
			callback:    commandRestart,
		},
		"save": {
			name:        "save",
			description: "saves data",
			callback:    commandSave,
		},
	}
	return commands
}

func cleanInput(text string) []string {
	var textWords []string
	//text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	firstPass := strings.Split(text, " ")

	for _, word := range firstPass {
		if word != "" {
			textWords = append(textWords, word)
		}
	}
	return textWords
}

func cleanInputAndStore(c *config, input string) {
	c.lastInput = cleanInput(input)
}

func commandHello(c *config) error {
	fmt.Println("Hello, World!")
	return nil
}

func commandHome(c *config) error {
	if c.location.currentNodeID == int(Home) {
		return fmt.Errorf("already at home")
	}
	err := c.location.ChangeNodeLoc("pharmacy")
	if err != nil {
		return err
	}

	commandClear(c)
	return nil
}

func commandExit(c *config) error {
	commandSave(c)
	fmt.Println("closing... goodbye!")
	c.rl.Close()
	os.Exit(0)
	return nil
}

func commandRestart(c *config) error {
	commandSave(c)
	commandClear(c)
	c.rl.Close()
	defer main()

	return nil
}

func commandRestartNoSave(c *config) error {
	commandClear(c)
	c.rl.Close()
	defer main()

	return nil
}

func commandSave(c *config) error {
	fmt.Println("saving data...")
	c.savePatientList()
	c.saveIgnoredOrdersList()
	c.savePrepullOrdersList()
	fmt.Println("save complete!")

	return nil
}

func commandClear(c *config) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

	default:
		//fmt.Print("\033[2J\033[H")
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	return nil
}

func commandSelect(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error too few arguments")
	}

	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "pt", "patient":
			err := commandSelectPatient(c)
			if err != nil {
				return err
			}
			return nil
		}

	}
	return nil
}

func commandAdd(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}
	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		// -- Adds for home like add patient
		switch firstArg {
		case "ignoredOrder":
			err := homeCommandAddIgnoredOrder(c)
			if err != nil {
				return err
			}

		case "order":
			err := homeCommandAddOrder(c)
			if err != nil {
				return err
			}

		case "prepullOrder":
			err := homeCommandAddPrepullOrder(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown item to add: %s not found", firstArg)
		}

	case PatientLoc:
		switch firstArg {
		case "order":
			err := patientCommandAddOrder(c)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("command does not exist for this location")
	}

	return nil
}

func commandGet(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}

	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "schedule":
			if len(c.lastInput) < 3 {
				return fmt.Errorf("error. missing schedule type")
			}
			secondArg := c.lastInput[2]
			switch secondArg {
			case "infusion", "-i", "inf":
				homeCommandGetScheduleInf(c)
				return nil

			case "clinic", "-c":
				//homeCommandGetScheduleClinic()
				return nil

			default:
				return fmt.Errorf("error. unknown schedule type: %s not found", secondArg)
			}

		case "next":
			if len(c.lastInput) < 3 {
				return fmt.Errorf("error. missing item")
			}
			secondArg := c.lastInput[2]

			switch secondArg {
			case "missingOrderPatient", "mop":
				err := homeCommandGetNextMissingOrderPatient(c)
				if err != nil {
					return err
				}
				return nil

			default:
				return fmt.Errorf("error. unknown next item: %s not found", secondArg)
			}

		case "prepullOrders":
			commandClear(c)
			err := homeCommandGetPrepullOrders(c)
			if err != nil {
				return err
			}

			return nil

		default:
			return fmt.Errorf("error. unknown argument: %s not found", firstArg)
		}
	default:
		return fmt.Errorf("get command cannot be used at the current location")
	}
}

func commandReview(c *config) error {
	// -- Error handling

	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {

	case Home:

		switch firstArg {

		case "moq", "missingOrdersQueue":
			err := homeCommandReviewMissingOrdersQueue(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown command: %s not a reviewable item", firstArg)
		}

	default:
		return fmt.Errorf("error. review command cannot be used from current node")
	}
	return nil
}

func commandMark(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}
	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "done":
			err := homeCommandMarkDone(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s is not a markable item", firstArg)
		}

	case PatientLoc:
		switch firstArg {
		case "order":
			err := patientCommandMarkOrder(c)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("error. mark command cannot be used from current node")
	}

	return nil
}

func commandRemove(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}
	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "order":
			err := homeCommandRemoveOrder(c)
			if err != nil {
				return err
			}

		case "saveData":
			err := homeCommandRemoveSaveData(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s is not a removable element", firstArg)
		}

	default:
		return fmt.Errorf("error. remove command cannot be used from current node")
	}

	return nil
}

func (c *config) commandLookup(input string) (cliCommand, error) {
	for _, c := range c.commands {
		if input == c.name {
			return c, nil
		}
	}
	return cliCommand{}, fmt.Errorf("unknown command")
}

func (c *config) CommandExe(input string) error {
	switch c.location.allNodes[c.location.currentNodeID].locType {
	case ReviewNode:
		missingOrdersREPL(c, input)

	default:
		cleanInputAndStore(c, input)
		if len(c.lastInput) == 0 {
			return nil
		}
		command, err := c.commandLookup(c.lastInput[0])
		if err != nil {
			return err
		}
		err = command.callback(c)
		if err != nil {
			return err
		}

	}
	return nil
}
