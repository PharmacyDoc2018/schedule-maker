package main

import (
	"fmt"
	"strings"
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

	c.AddOrderQuick(mrn, order)

	fmt.Printf("order: %s added for %s\n", order, ptName)

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
