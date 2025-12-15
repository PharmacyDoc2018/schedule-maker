package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

func initREPL() *config {
	godotenv.Load(".env")
	pathToSch := os.Getenv("SCH_PATH")
	pathToSave := os.Getenv("SAVE_PATH")
	pathToPrepullOrders := os.Getenv("PREPULL_ORDERS_PATH")
	pathToIgnoredOrders := os.Getenv("IGNORED_ORDERS_PATH")
	pathToPtSupplyOrders := os.Getenv("PT_SUPPLY_ORDERS_PATH")

	config := &config{
		pathToSch:            pathToSch,
		pathToSave:           pathToSave,
		pathToIgnoredOrders:  pathToIgnoredOrders,
		pathToPrepullOrders:  pathToPrepullOrders,
		pathToPtSupplyOrders: pathToPtSupplyOrders,
	}

	config.commands = getCommands()

	config.PatientList.Map = map[string]Patient{}
	config.patientNameMap = map[string]struct{}{}
	config.IgnoredOrders.Map = map[string]struct{}{}
	config.PtSupplyOrders.Map = map[string]map[string]struct{}{}

	nodeMap := make(map[int]*LocationNode)
	nodeMap[0] = &LocationNode{
		id:       0,
		name:     "pharmacy",
		locType:  Home,
		parentID: -1,
	}
	config.location = Location{
		allNodes:      nodeMap,
		currentNodeID: 0,
	}

	config.lastSave = time.Now()

	config.readlineCompleterMap = map[int]*readline.PrefixCompleter{}

	return config
}

func (c *config) readlineSetup() *readline.Instance {
	completerMode := make(map[int]*readline.PrefixCompleter)

	completerMode[int(Home)] = readline.NewPrefixCompleter(
		readline.PcItem("select",
			readline.PcItem("pt",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("patient",
				readline.PcItemDynamic(c.getPatientArgs),
			),
		),
		readline.PcItem("exit"),
		readline.PcItem("get",
			readline.PcItem("schedule",
				readline.PcItem("infusion",
					readline.PcItem("allOrders"),
					readline.PcItem("-ao"),
				),
				readline.PcItem("-i",
					readline.PcItem("allOrders"),
					readline.PcItem("-ao"),
				),
				readline.PcItem("inf",
					readline.PcItem("allOrders"),
					readline.PcItem("-ao"),
				),
			),
			readline.PcItem("prepullOrders"),
			readline.PcItem("next",
				readline.PcItem("mop"),
				readline.PcItem("missingOrderPatient"),
			),
			readline.PcItem("orders",
				readline.PcItem("-a"),
			),
			readline.PcItem("ptSupplied",
				readline.PcItem("-a"),
			),
		),
		readline.PcItem("clear"),
		readline.PcItem("review",
			readline.PcItem("moq"),
			readline.PcItem("missingOrdersQueue"),
		),
		readline.PcItem("add",
			readline.PcItem("ignoredOrder"),
			readline.PcItem("order",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("prepullOrder"),
			readline.PcItem("patient"),
		),
		readline.PcItem("mark",
			readline.PcItem("done",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("order",
				readline.PcItemDynamic(c.getPatientArgs,
					readline.PcItemDynamic(c.getPatientOrders),
				),
			),
			readline.PcItem("ptSupplied",
				readline.PcItemDynamic(c.getPatientArgs),
			),
		),
		readline.PcItem("remove",
			readline.PcItem("order",
				readline.PcItemDynamic(c.getPatientArgs,
					readline.PcItemDynamic(c.getPatientOrders),
				),
			),
			readline.PcItem("ignoredOrder"),
			readline.PcItem("saveData"),
			readline.PcItem("ptSupplied",
				readline.PcItemDynamic(c.getPatientArgs,
					readline.PcItemDynamic(c.getPatientOrders),
				),
			),
			readline.PcItem("prepullOrder"),
			readline.PcItem("ptList",
				readline.PcItemDynamic(readlinePcDynamicItemHelper(c.PatientLists.GetDates)),
			),
			readline.PcItem("patient",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("done",
				readline.PcItemDynamic(c.getPatientArgs),
			),
		),
		readline.PcItem("reset"),
		readline.PcItem("save"),
		readline.PcItem("change",
			readline.PcItem("apptTimeInf",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("appointmentTimeInfusion",
				readline.PcItemDynamic(c.getPatientArgs),
			),
			readline.PcItem("ptList",
				readline.PcItemDynamic(readlinePcDynamicItemHelper(c.PatientLists.GetDates)),
			),
		),
		readline.PcItem("list",
			readline.PcItem("ignoredOrders"),
			readline.PcItem("prepullOrders"),
			readline.PcItem("ptLists"),
		),
		readline.PcItem("load",
			readline.PcItem("excelData"),
		),
	)

	completerMode[int(PatientLoc)] = readline.NewPrefixCompleter(
		readline.PcItem("add",
			readline.PcItem("order"),
		),
		readline.PcItem("home"),
		readline.PcItem("exit"),
		readline.PcItem("mark",
			readline.PcItem("order",
				readline.PcItemDynamic(c.GetPatientOrdersFromLoc),
			),
			readline.PcItem("ptSupplied"),
			readline.PcItem("done"),
		),
		readline.PcItem("remove",
			readline.PcItem("order",
				readline.PcItemDynamic(c.GetPatientOrdersFromLoc),
			),
			readline.PcItem("ptSupplied",
				readline.PcItemDynamic(c.GetPatientOrdersFromLoc),
			),
			readline.PcItem("done"),
		),
		readline.PcItem("change",
			readline.PcItem("apptTimeInf"),
			readline.PcItem("appointmentTimeInfusion"),
		),
	)

	completerMode[int(ReviewNode)] = readline.NewPrefixCompleter(
	//
	)

	c.readlineCompleterMap = completerMode

	rl, _ := readline.NewEx(&readline.Config{
		Prompt:       c.location.Path(),
		AutoComplete: completerMode[int(Home)],
	})
	c.readlineConfig = rl.Config.Clone()

	return rl

}

func readlinePcDynamicItemHelper(f func() []string) func(string) []string {
	return func(string) []string {
		return f()
	}
}

func (c *config) getPatientArgs(input string) []string {
	var patients []string
	for _, val := range c.PatientList.Map {
		patients = append(patients, val.Name)
	}

	return patients
}

func (c *config) getPatientOrders(input string) []string {
	orders := []string{}
	inputParts := strings.Split(input, " ")[2:]

	ptName := ""
	for i := 0; i < len(inputParts); i++ {
		if _, ok := c.patientNameMap[strings.Join(inputParts[:i], " ")]; ok {
			ptName = strings.Join(inputParts[:i], " ")
			break
		}
	}
	if ptName == "" {
		return orders
	}

	mrn := ""
	for key, val := range c.PatientList.Map {
		if val.Name == ptName {
			mrn = key
			break
		}
	}
	if mrn == "" {
		return orders
	}

	for _, order := range c.PatientList.Map[mrn].Orders {
		orders = append(orders, order)
	}

	return orders
}

func (c *config) GetPatientOrdersFromLoc(input string) []string {
	orders := []string{}

	mrn := c.location.allNodes[c.location.currentNodeID].name
	for _, order := range c.PatientList.Map[mrn].Orders {
		orders = append(orders, order)
	}

	return orders
}

func (c *config) resetPrefixCompleterMode() {
	switch c.location.allNodes[c.location.currentNodeID].locType {
	case Home:
		c.readlineConfig.AutoComplete = c.readlineCompleterMap[int(Home)]

	case PatientLoc:
		c.readlineConfig.AutoComplete = c.readlineCompleterMap[int(PatientLoc)]

	case ReviewNode:
		c.readlineConfig.AutoComplete = c.readlineCompleterMap[int(ReviewNode)]

	}
}

func (c *config) readlineLoopStartPreprocess() {
	switch c.location.allNodes[c.location.currentNodeID].locType {
	case ReviewNode:

		switch c.location.allNodes[c.location.currentNodeID].name {
		case "Missing Orders Queue":
			mrn, err := c.missingOrders.NextPatient()
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println("returning to home location...")
				c.location.ChangeNodeLoc("pharmacy")
				c.resetPrefixCompleterMode()
				c.rl.SetConfig(c.readlineConfig)
				c.rl.SetPrompt(c.location.Path())
				return
			}

			pt := c.PatientList.Map[mrn].Name
			apptTime := func() string {
				for appt, apptTime := range c.PatientList.Map[mrn].AppointmentTimes {
					if strings.Contains(appt, infusionAppointmentTag) {
						return apptTime.Format(timeFormat)
					}
				}
				return ""
			}()

			fmt.Printf("Current Patient: %s - %s (%s). Adding orders:\n", apptTime, pt, mrn)
			if len(c.PatientList.Map[mrn].Orders) > 0 {
				fmt.Println("Current Orders:")
				for _, order := range c.PatientList.Map[mrn].Orders {
					fmt.Println(order)
				}
				fmt.Println()
			}
		}

	case Home:
		var infusionPatientNum int
		var remainingPatientNum int
		for mrn := range c.PatientList.Map {
			for appt := range c.PatientList.Map[mrn].AppointmentTimes {
				if strings.Contains(appt, infusionAppointmentTag) {
					infusionPatientNum++
					if !c.PatientList.Map[mrn].VisitComplete {
						remainingPatientNum++
					}
					break
				}
			}
		}

		displayMessage := ""

		if !isSameDay(c.PatientList.Date, time.Now()) {
			displayMessage += c.PatientList.Date.Format(dateFormat) + ": "
		}

		displayMessage += fmt.Sprintf("Infusion Patients(%d)", infusionPatientNum)

		if c.missingOrders.Len() > 0 {
			displayMessage += fmt.Sprintf(". Missing Orders(%d)", c.missingOrders.Len())
		}

		if remainingPatientNum > 0 {
			displayMessage += fmt.Sprintf(". Remaining Patients(%d)", remainingPatientNum)
		} else {
			displayMessage += ". All patients completed"
		}

		displayMessage += ":"
		fmt.Println(displayMessage)

	case PatientLoc:
		commandClear(c)
		mrn := c.location.allNodes[c.location.currentNodeID].name
		pt := c.PatientList.Map[mrn]
		fmt.Printf("Selected Patient: %s (%s)\n", pt.Name, pt.Mrn)

		if c.PatientList.Map[mrn].VisitComplete {
			fmt.Println("-- Visit Completed --")
		}
		fmt.Println()

		if len(c.PatientList.Map[mrn].AppointmentTimes) == 0 {
			fmt.Println("Appointments: None")
		} else {
			fmt.Println("Appointments:")
			apptSlices := []string{}
			for key := range c.PatientList.Map[mrn].AppointmentTimes {
				apptSlices = append(apptSlices, key)
			}

			sort.Slice(apptSlices, func(i, j int) bool {
				return c.PatientList.Map[mrn].AppointmentTimes[apptSlices[i]].Before(c.PatientList.Map[mrn].AppointmentTimes[apptSlices[j]])
			})

			for _, appt := range apptSlices {
				fmt.Printf("  %s: %s\n", appt, c.PatientList.Map[mrn].AppointmentTimes[appt].Format(timeFormat))
			}
			fmt.Println()
		}

		if len(c.PatientList.Map[mrn].Orders) == 0 {
			fmt.Println("Current Orders: None")
		} else {
			currentOrders := []string{}
			ignoredOrders := []string{}
			for _, order := range c.PatientList.Map[mrn].Orders {
				if c.IgnoredOrders.Exists(order) {
					ignoredOrders = append(ignoredOrders, order)
				} else if c.PtSupplyOrders.IsPatientSupplied(mrn, order) {
					currentOrders = append(currentOrders, fmt.Sprintf("[Pt Supplied] %s [Pt Supplied]", order))
				} else {
					currentOrders = append(currentOrders, order)
				}
			}
			fmt.Println("Current Orders:")
			for _, order := range currentOrders {
				fmt.Println(" ", order)
			}
			fmt.Println()
			fmt.Println("Other Orders:")
			for _, order := range ignoredOrders {
				fmt.Println(" ", order)
			}

		}
		fmt.Println()
	}
}

func (c *config) FindPatientInInput(start int) (mrn string, err error) {
	i := start + 1
	for i <= len(c.lastInput) {
		if _, ok := c.patientNameMap[strings.Join(c.lastInput[start:i], " ")]; ok {
			ptName := strings.Join(c.lastInput[2:i], " ")
			for key, val := range c.PatientList.Map {
				if val.Name == ptName {
					return key, nil
				}
			}
		}
		i++
	}
	return "", fmt.Errorf("error. patient not found")

}

func (c *config) FindPatientItemInInput(start int, itemType string) (mrn, ptName, item string, err error) {
	mrn, err = c.FindPatientInInput(start)
	if err != nil {
		return "", "", "", err
	}

	ptName = c.PatientList.Map[mrn].Name
	commandPatientLen := len(strings.Split(ptName, " ")) + start
	if len(c.lastInput) == commandPatientLen {
		return "", "", "", fmt.Errorf("error. missing %s argument", itemType)
	}

	item = strings.Join(c.lastInput[commandPatientLen:], " ")

	return mrn, ptName, item, nil
}

func (c *config) AutoSave() {
	const autosaveInterval = 5 * time.Minute
	if time.Since(c.lastSave) >= autosaveInterval {
		commandSave(c)
	}

}
