package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

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

	config.PatientList = map[string]Patient{}
	config.patientNameMap = map[string]struct{}{}
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
			readline.PcItem("clear"),
		),
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
		),
		readline.PcItem("mark",
			readline.PcItem("done",
				readline.PcItemDynamic(c.getPatientArgs),
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
		),
		readline.PcItem("remove",
			readline.PcItem("order",
				readline.PcItemDynamic(c.GetPatientOrdersFromLoc),
			),
			readline.PcItem("ptSupplied",
				readline.PcItemDynamic(c.GetPatientOrdersFromLoc),
			),
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

func (c *config) getPatientArgs(input string) []string {
	var patients []string
	for _, val := range c.PatientList {
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
	for key, val := range c.PatientList {
		if val.Name == ptName {
			mrn = key
			break
		}
	}
	if mrn == "" {
		return orders
	}

	for _, order := range c.PatientList[mrn].Orders {
		orders = append(orders, order)
	}

	return orders
}

func (c *config) GetPatientOrdersFromLoc(input string) []string {
	orders := []string{}

	mrn := c.location.allNodes[c.location.currentNodeID].name
	for _, order := range c.PatientList[mrn].Orders {
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

			pt := c.PatientList[mrn].Name
			apptTime := func() string {
				for appt, apptTime := range c.PatientList[mrn].AppointmentTimes {
					if strings.Contains(appt, infusionAppointmentTag) {
						return apptTime.Format(timeFormat)
					}
				}
				return ""
			}()

			fmt.Printf("Current Patient: %s - %s (%s). Adding orders:\n", apptTime, pt, mrn)
			if len(c.PatientList[mrn].Orders) > 0 {
				fmt.Println("Current Orders:")
				for _, order := range c.PatientList[mrn].Orders {
					fmt.Println(order)
				}
				fmt.Println()
			}
		}

	case Home:
		var infusionPatientNum int
		for mrn := range c.PatientList {
			for appt := range c.PatientList[mrn].AppointmentTimes {
				if strings.Contains(appt, infusionAppointmentTag) {
					infusionPatientNum++
					break
				}
			}
		}
		fmt.Printf("Infusion Patients(%d). Missing Orders(%d):\n", infusionPatientNum, c.missingOrders.Len())

	case PatientLoc:
		commandClear(c)
		mrn := c.location.allNodes[c.location.currentNodeID].name
		pt := c.PatientList[mrn]
		fmt.Printf("Selected Patient: %s (%s)\n\n", pt.Name, pt.Mrn)

		if len(c.PatientList[mrn].AppointmentTimes) == 0 {
			fmt.Println("Appointments: None")
		} else {
			fmt.Println("Appointments:")
			apptSlices := []string{}
			for key := range c.PatientList[mrn].AppointmentTimes {
				apptSlices = append(apptSlices, key)
			}

			sort.Slice(apptSlices, func(i, j int) bool {
				return c.PatientList[mrn].AppointmentTimes[apptSlices[i]].Before(c.PatientList[mrn].AppointmentTimes[apptSlices[j]])
			})

			for _, appt := range apptSlices {
				fmt.Printf("  %s: %s\n", appt, c.PatientList[mrn].AppointmentTimes[appt].Format(timeFormat))
			}
			fmt.Println()
		}

		if len(c.PatientList[mrn].Orders) == 0 {
			fmt.Println("Current Orders: None")
		} else {
			fmt.Println("Current Orders:")
			for _, order := range c.PatientList[mrn].Orders {
				if c.PtSupplyOrders.IsPatientSupplied(mrn, order) {
					fmt.Println(" ", "[Pt Supplied]", order, "[Pt Supplied]")
				} else {
					fmt.Println(" ", order)
				}
			}
		}
		fmt.Println()
	}
}

func (c *config) FindPatientInInput(start int) (mrn string, err error) {
	i := start + 1
	for i < len(c.patientNameMap) {
		if _, ok := c.patientNameMap[strings.Join(c.lastInput[start:i], " ")]; ok {
			ptName := strings.Join(c.lastInput[2:i], " ")
			for key, val := range c.PatientList {
				if val.Name == ptName {
					return key, nil
				}
			}
		}
		i++
	}
	return "", fmt.Errorf("error. patient not found")

}
