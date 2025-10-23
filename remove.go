package main

import (
	"fmt"
	"os"
	"strings"
)

func homeCommandRemoveOrder(c *config) error {
	//
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments.\nExpected format: remove order [pt name] [order name]")
	}

	mrn, err := c.FindPatientInInput(2)
	if err != nil {
		return err
	}

	ptName := c.PatientList[mrn].Name
	if len(ptName)+2 == len(c.lastInput) {
		return fmt.Errorf("error. no order entered")
	}

	ptNameLen := len(strings.Split(ptName, " "))
	order := strings.Join(c.lastInput[ptNameLen+2:], " ")
	err = c.RemoveOrder(mrn, order)
	if err != nil {
		return err
	}

	fmt.Printf("removed order %s from %s\n", order, ptName)

	return nil
}

func homeCommandRemoveSaveData(c *config) error {
	err := os.Remove(c.pathToSave)
	if err != nil {
		return err
	}
	commandRestartNoSave(c)

	return nil
}

func homeCommandRemoveIgnoredOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing order argument")
	}

	order := strings.Join(c.lastInput[2:], " ")
	storedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	oldList := c.IgnoredOrders.List
	newList := []string{}
	for i, item := range oldList {
		if storedOrder != item {
			newList = append(newList, item)
		} else {
			newList = append(newList, oldList[i+1:]...)
			c.IgnoredOrders.List = newList
			oldList = nil
			fmt.Printf("%s removed from Ignored Orders List\n", order)
			return nil
		}
	}
	return fmt.Errorf("error. order: %s not found in Ignored Orders List", order)

}

func patientCommandRemoveOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. too few arguments.\nExpected format: remove order [pt name] [order name]")
	}

	order := strings.Join(c.lastInput[2:], " ")
	mrn := c.location.allNodes[c.location.currentNodeID].name
	patient := c.PatientList[mrn]

	for key, val := range patient.Orders {
		if val == order {
			delete(patient.Orders, key)
			c.PatientList[mrn] = patient
			return nil
		}
	}

	return fmt.Errorf("error. order: %s not found", order)
}

func patientCommandRemovePtSupplied(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing medication argument")
	}

	order := strings.Join(c.lastInput[2:], " ")
	mrn := c.location.allNodes[c.location.currentNodeID].name

	err := c.PtSupplyOrders.RemoveOrder(mrn, order)
	if err != nil {
		return err
	}

	return nil

}
