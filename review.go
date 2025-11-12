package main

import "fmt"

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
		input = c.OrderPreprocessing(input) //-- replaces dot phrases
		c.AddOrderQuick(mrn, input)
		fmt.Println("order added: ", input)
		commandClear(c)

	}

}
