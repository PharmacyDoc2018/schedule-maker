package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
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
		"..": {
			name:        "..",
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
		"change": {
			name:        "change",
			description: "change the value of an item. i.e. appointment time",
			callback:    commandChange,
		},
		"list": {
			name:        "list",
			description: "prints elements in a list",
			callback:    commandList,
		},
		"load": {
			name:        "load",
			description: "loads data from a save file",
			callback:    commandLoad,
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
	err := c.savePatientLists()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = c.saveIgnoredOrdersList()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = c.savePrepullOrdersList()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = c.savePtSupplyOrderList()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = c.saveProviders()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("save complete!")
	c.lastSave = time.Now()

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

func commandChange(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}

	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "apptTimeInf", "appointmentTimeInfusion":
			err := homeCommandChangeApptTimeInf(c)
			if err != nil {
				return err
			}

		case "ptList":
			err := homeCommandChangePatientList(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s not a changeable item", firstArg)
		}

	case PatientLoc:
		switch firstArg {
		case "apptTimeInf", "appointmentTimeInfusion":
			err := patientCommandChangeApptTimeInf(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s not a changeable item", firstArg)
		}
	default:
		return fmt.Errorf("error. cannot use the change command at current location")
	}

	return nil
}

func commandSelect(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
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
			c.missingOrders = c.PatientList.FindMissingInfusionOrders()

		case "prepullOrder":
			err := homeCommandAddPrepullOrder(c)
			if err != nil {
				return err
			}

		case "patient":
			err := homeCommandAddPatient(c)
			if err != nil {
				return err
			}
			c.missingOrders = c.PatientList.FindMissingInfusionOrders()
			c.createPatientNameMap()
			commandSave(c)

		case "provider":
			err := homeCommandAddProvider(c)
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
			c.missingOrders = c.PatientList.FindMissingInfusionOrders()
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
				err := homeCommandGetScheduleInf(c)
				if err != nil {
					return err
				}

			case "clinic", "-c":
				err := homeCommandGetScheduleClinic(c)
				if err != nil {
					return err
				}

			case "nurse":
				err := homeCommandGetScheduleNurse(c)
				if err != nil {
					return err
				}

			default:
				concatRemainingArgs := strings.Join(c.lastInput[2:], " ")
				name, err := c.FindProviderInInput(2)
				if err != nil {
					return err
				}
				if c.Providers.Exists(name) {
					err := homeCommandGetScheduleProvider(c, concatRemainingArgs)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("error. unknown schedule type: %s not found", secondArg)
				}
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

			default:
				return fmt.Errorf("error. unknown next item: %s not found", secondArg)
			}

		case "prepullOrders":
			commandClear(c)
			err := homeCommandGetPrepullOrders(c)
			if err != nil {
				return err
			}

		case "orders":
			commandClear(c)
			err := homeCommandGetOrders(c)
			if err != nil {
				return err
			}

		case "ptSupplied":
			commandClear(c)
			err := homeCommandGetPtSupplied(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. unknown argument: %s not found", firstArg)
		}
	default:
		return fmt.Errorf("get command cannot be used at the current location")
	}

	return nil
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

		case "ptSupplied":
			err := homeCommandMarkPtSupplied(c)
			if err != nil {
				return err
			}

		case "order":
			err := homeCommandMarkOrder(c)
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

		case "ptSupplied":
			err := patientCommandMarkPtSupplied(c)
			if err != nil {
				return err
			}

		case "done":
			err := patientCommandMarkDone(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s is not a markable item", firstArg)
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
			c.missingOrders = c.PatientList.FindMissingInfusionOrders()

		case "saveData":
			err := homeCommandRemoveSaveData(c)
			if err != nil {
				return err
			}

		case "ignoredOrder":
			err := homeCommandRemoveIgnoredOrder(c)
			if err != nil {
				return err
			}

		case "ptSupplied":
			err := homeCommandRemovePtSupplied(c)
			if err != nil {
				return err
			}

		case "prepullOrder":
			err := homeCommandRemovePrepullOrder(c)
			if err != nil {
				return err
			}

		case "ptList":
			err := homeCommandRemovePatientList(c)
			if err != nil {
				return err
			}

		case "patient":
			err := homeCommandRemovePatient(c)
			if err != nil {
				return err
			}
			c.missingOrders = c.PatientList.FindMissingInfusionOrders()
			c.createPatientNameMap()
			commandSave(c)

		case "done":
			err := homeCommandRemoveDone(c)
			if err != nil {
				return err
			}

		case "provider":
			err := homeCommandRemoveProvider(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s is not a removable element", firstArg)
		}

	case PatientLoc:
		switch firstArg {
		case "order":
			err := patientCommandRemoveOrder(c)
			if err != nil {
				return err
			}
			c.missingOrders = c.PatientList.FindMissingInfusionOrders()

		case "ptSupplied":
			err := patientCommandRemovePtSupplied(c)
			if err != nil {
				return err
			}

		case "done":
			err := patientCommandRemoveDone(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s not a removable element", firstArg)
		}

	default:
		return fmt.Errorf("error. remove command cannot be used from current node")
	}

	return nil
}

func commandList(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}

	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "ignoredOrders":
			err := homeCommandListIgnoredOrders(c)
			if err != nil {
				return err
			}

		case "prepullOrders":
			err := homeCommandListPrepullOrders(c)
			if err != nil {
				return err
			}

		case "ptLists":
			err := homeCommandListPatientLists(c)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("error. %s not a listable item", firstArg)
		}

	default:
		return fmt.Errorf("error. list command cannot be used from current node")

	}

	return nil
}

func commandLoad(c *config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}

	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		switch firstArg {
		case "excelData":
			err := homeCommandLoadExcelData(c)
			if err != nil {
				return err
			}
			commandSave(c)

		default:
			return fmt.Errorf("error. %s not a loadable item", firstArg)
		}

	default:
		return fmt.Errorf("error. load command cannot be used from current node")
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

		c.AutoSave()

	}
	return nil
}
