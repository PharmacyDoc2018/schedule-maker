package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func homeCommandAddIgnoredOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing order argument")
	}

	order := strings.Join(c.lastInput[2:], " ")
	err := c.IgnoredOrders.Add(order)
	if err != nil {
		return err
	}

	fmt.Printf("order added: %s will be ignored\n", order)

	return nil
}

func homeCommandAddPrepullOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing order argument")
	}

	order := strings.Join(c.lastInput[2:], " ")
	err := c.PrepullOrders.Add(order)
	if err != nil {
		return err
	}

	fmt.Printf("order: %s added to Prepull Orders list\n", order)

	return nil
}

func homeCommandAddOrder(c *config) error {
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments.\nExpected format: add order [pt name] [order name]")
	}

	mrn, ptName, order, err := c.FindPatientItemInInput(2, "order")
	if err != nil {
		return err
	}

	order = c.OrderPreprocessing(order) //-- swaps out dot phrases
	c.PatientList.AddOrderQuick(mrn, order)

	fmt.Printf("order: %s added for %s\n", order, ptName)

	return nil
}

func homeCommandAddPatient(c *config) error {
	addPatientScanner := bufio.NewScanner(os.Stdin)
	prompts := []string{
		"MRN: ",
		"Name: ",
		"Appointment Time: ",
		"Apptintment Type: ",
	}
	inputs := []string{}

	for i := 0; i <= 3; i++ {
		fmt.Print(prompts[i])
		addPatientScanner.Scan()
		input := addPatientScanner.Text()
		inputs = append(inputs, input)
	}

	mrn := inputs[0]
	name := inputs[1]
	apptTimeString := inputs[2]
	apptType := inputs[3]

	err := c.PatientList.addPatient(mrn, name)
	if err != nil {
		return err
	}

	scheduleType := func(apptType string) string {
		switch apptType {
		case "inf", "INF", "infusion", "Infusion", "AUBL INF":
			return infusionAppointmentTag

		default:
			return "AUBL CANC"
		}
	}(apptType)

	err = c.PatientList.addAppointment(mrn, scheduleType, time.Now().Format(dateFormat), apptTimeString)
	if err != nil {
		fmt.Printf("error. patient added, but unable to add appointment: %s\n", err.Error())
		return nil
	}

	fmt.Printf("%s successfully added to the schedule\n", name)
	return nil
}

func patientCommandAddOrder(c *config) error {
	order := strings.Join(c.lastInput[2:], " ")
	order = c.OrderPreprocessing(order) //-- swaps out dot phrases
	mrn := c.location.allNodes[c.location.currentNodeID].name
	c.PatientList.AddOrderQuick(mrn, order)
	fmt.Println("order added: ", order)

	err := c.missingOrders.RemovePatient(mrn)
	if err == nil {
		fmt.Println("patient removed from missing orders list")
	}

	return nil
}
