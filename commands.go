package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
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
	fmt.Println("saving schedule...")
	c.savePatientList()
	fmt.Println("schedule saved!")
	c.saveIgnoredOrdersList()
	fmt.Println("closing... goodbye!")
	c.rl.Close()
	os.Exit(0)
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

func commandSelectPatient(c *config) error {
	var pt string
	if len(c.lastInput) == 3 {
		pt = c.lastInput[2]
	} else {
		pt = strings.Join(c.lastInput[2:], " ")
	}

	if _, ok := c.PatientList[pt]; ok {
		err := c.location.SelectPatientNode(pt)
		if err != nil {
			return err
		}
		return nil

	} else {
		for key, val := range c.PatientList {
			if pt == val.Name {
				fmt.Println("Found patient with name. MRN is", key)
				fmt.Println()
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
		switch firstArg {
		case "ignoredOrder":
			err := homeCommandAddIgnoredOrder(c)
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

func homeCommandAddIgnoredOrder(c *config) error {
	order := strings.Join(c.lastInput[2:], " ")
	storedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	for _, item := range c.IgnoredOrders.List {
		if storedOrder == item {
			return fmt.Errorf("entry already exists on the Ignored Orders list")
		}
	}

	c.IgnoredOrders.List = append(c.IgnoredOrders.List, storedOrder)
	fmt.Printf("order added: %s will be ignored\n", order)

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

	case "next":

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

	default:
		return fmt.Errorf("error. unknown argument: %s not found", firstArg)
	}
}

func homeCommandGetScheduleInf(c *config) {
	schedule := Schedule{
		colSpaceBuffer: 2,
	}

	type infAppt struct {
		time   string
		mrn    string
		name   string
		orders []string
	}

	const timeFormat = "3:04 PM"

	infApptSlices := []infAppt{}
	for _, patient := range c.PatientList {
		for appt, apptTime := range patient.AppointmentTimes {
			if strings.Contains(appt, infusionAppointmentTag) {
				ordersSlice := []string{}
				for _, order := range patient.Orders {
					ordersSlice = append(ordersSlice, order)
				}
				infApptSlices = append(infApptSlices, infAppt{
					time:   apptTime.Format(timeFormat),
					mrn:    patient.Mrn,
					name:   patient.Name,
					orders: ordersSlice,
				})
				break
			}
		}
	}

	sort.Slice(infApptSlices, func(i, j int) bool {
		a, _ := time.Parse(timeFormat, infApptSlices[i].time)
		b, _ := time.Parse(timeFormat, infApptSlices[j].time)
		return a.Before(b)
	})

	for _, appt := range infApptSlices {
		if len(appt.orders) > 0 {
			schedule.table = append(schedule.table, []string{
				appt.time,
				appt.mrn,
				appt.name,
				appt.orders[0],
			})
			for _, order := range appt.orders[1:] {
				schedule.table = append(schedule.table, []string{
					"",
					"",
					"",
					order,
				})
			}
		} else {
			schedule.table = append(schedule.table, []string{
				appt.time,
				appt.mrn,
				appt.name,
				"",
			})
		}

	}
	//commandClear(c)
	schedule.Print(c, []string{"default"})
}

func homeCommandGetNextMissingOrderPatient(c *config) error {
	mrn, err := c.missingOrders.NextPatient()
	if err != nil {
		return err
	}

	err = c.location.SelectPatientNode(mrn)
	if err != nil {
		return err
	}

	pt := c.PatientList[mrn].Name
	fmt.Printf("next patient with missing orders: %s (%s)\n", pt, mrn)

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
		return fmt.Errorf("error. Review command cannot be used from current node")
	}
	return nil
}

func homeCommandReviewMissingOrdersQueue(c *config) error {
	// since the review node changes the REPL entirely, REPL logic handled by
	// missingOrdersREPL()
	err := c.location.SelectReviewNode("Missing Orders Queue")
	if err != nil {
		return err
	}
	return nil
}

func missingOrdersREPL(c *config, input string) {
	mrn, _ := c.missingOrders.NextPatient()
	switch input {
	case "":
		fmt.Println("loading next patient...")
		c.missingOrders.PopPatient()
		if len(c.PatientList[mrn].Orders) == 0 {
			c.missingOrders.AddPatient(mrn)
		}

	case "q", "quit":
		fmt.Println("exiting Missing Orders Queue...")
		err := c.location.ChangeNodeLoc("pharmacy")
		if err != nil {
			fmt.Println(err.Error())
		}

	default:
		c.AddOrderQuick(mrn, input)
		fmt.Println("order added: ", input)
		commandClear(c)

	}

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
