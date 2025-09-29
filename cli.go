package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

func initREPL() *config {
	godotenv.Load(".env")
	pathToSch := os.Getenv("SCH_PATH")
	pathToSave := os.Getenv("SAVE_PATH")
	pathToIgnoredOrders := os.Getenv("IGNORED_ORDERS_PATH")

	config := &config{
		pathToSch:           pathToSch,
		pathToSave:          pathToSave,
		pathToIgnoredOrders: pathToIgnoredOrders,
	}

	config.commands = getCommands()

	config.PatientList = map[string]Patient{}

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
				readline.PcItem("infusion"),
				readline.PcItem("-i"),
			),
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
		),
	)

	completerMode[int(PatientLoc)] = readline.NewPrefixCompleter(
		readline.PcItem("add",
			readline.PcItem("order"),
		),
		readline.PcItem("home"),
		readline.PcItem("exit"),
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
						return apptTime.Format("15:04")
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
	}
}
