package main

import (
	"fmt"
	"os"
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

	return nil
}

func commandExit(c *config) error {
	// -- will need to save data once implemented
	fmt.Println("closing... goodbye!")
	c.rl.Close()
	os.Exit(0)
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

func commandSelectPatient(c *config) error {
	var pt string
	if len(c.lastInput) == 3 {
		pt = c.lastInput[2]
	} else {
		pt = strings.Join(c.lastInput[2:], " ")
	}

	if _, ok := c.patientList[pt]; ok {
		err := c.location.SelectPatientNode(pt)
		if err != nil {
			return err
		}
		return nil

	} else {
		for key, val := range c.patientList {
			if pt == val.name {
				fmt.Println("Found patient with name. MRN is", key)
				err := c.location.SelectPatientNode(key)
				if err != nil {
					return err
				}

			}
		}
		return nil
	}
}

func commandAdd(c *config) error {
	firstArg := c.lastInput[1]

	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		// -- Adds for home like add patient

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

func patientCommandAddOrder(c *config) error {
	order := strings.Join(c.lastInput[2:], " ")
	mrn := c.location.allNodes[c.location.currentNodeID].name
	c.AddOrderQuick(mrn, order)
	fmt.Println("order added: ", order)

	err := c.missingOrders.RemovePatient(mrn)
	if err == nil {
		fmt.Println("patient removed from missing orders list")
	}

	return nil
}

func commandGet(c *config) error {
	// --  input error handling

	firstArg := c.lastInput[1]
	secondArg := c.lastInput[2]

	switch firstArg {
	case "schedule":

		switch secondArg {
		case "infusion", "-i":
			homeCommandGetScheduleInf(c)
			return nil

		case "clinic", "-c":
			//homeCommandGetScheduleClinic()
			return nil

		default:
			return fmt.Errorf("error. unknown schedule type: %s not found", secondArg)
		}

	default:
		return fmt.Errorf("error. unknown argument: %s not found", firstArg)
	}
}

func homeCommandGetScheduleInf(c *config) {
	const infusionAppointmentTag = "AUBL INF"
	schedule := Schedule{
		colSpaceBuffer: 2,
	}

	for _, patient := range c.patientList {
		for appt, apptTime := range patient.appointmentTimes {
			if strings.Contains(appt, infusionAppointmentTag) {
				if len(patient.orders) > 0 {
					ordersSlice := []string{}
					for _, order := range patient.orders {
						ordersSlice = append(ordersSlice, order)
					}
					schedule.table = append(schedule.table, []string{
						apptTime.Format("15:04"),
						patient.mrn,
						patient.name,
						ordersSlice[0],
					})
					for _, order := range ordersSlice[1:] {
						schedule.table = append(schedule.table, []string{
							"",
							"",
							"",
							order,
						})
					}
				} else {
					schedule.table = append(schedule.table, []string{
						apptTime.Format("15:04"),
						patient.mrn,
						patient.name,
						"",
					})
				}
				break
			}
		}
	}

	schedule.Print()
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
	cleanInputAndStore(c, input)
	command, err := c.commandLookup(c.lastInput[0])
	if err != nil {
		return err
	}
	err = command.callback(c)
	if err != nil {
		return err
	}
	return nil
}
