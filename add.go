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

func homeCommandAddOrder(c *config) error {
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments.\nExpected format: add order [pt name] [order name]")
	}

	ptList := []string{}
	for _, val := range c.PatientList {
		ptList = append(ptList, val.Name)
	}

	ptNameMap := make(map[string]struct{}, len(ptList))
	for _, name := range ptList {
		ptNameMap[name] = struct{}{}
	}
	ptList = nil

	ptName, err := func() (string, error) {
		i := 3
		for i < len(ptNameMap) {
			if _, ok := ptNameMap[strings.Join(c.lastInput[2:i], " ")]; ok {
				return strings.Join(c.lastInput[2:i], " "), nil
			}
			i++
		}
		return "", fmt.Errorf("error. patient not found")
	}()
	if err != nil {
		return err
	}

	mrn := ""
	for key, val := range c.PatientList {
		if val.Name == ptName {
			mrn = key
			break
		}
	}
	if mrn == "" {
		return fmt.Errorf("error. mrn not found for %s", ptName)
	}

	if len(ptName)+2 == len(c.lastInput) {
		return fmt.Errorf("error. no order entered")
	}

	ptNameLen := len(strings.Split(ptName, " "))
	order := strings.Join(c.lastInput[ptNameLen+2:], " ")
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
