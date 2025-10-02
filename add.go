package main

import (
	"fmt"
	"strings"
)

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
