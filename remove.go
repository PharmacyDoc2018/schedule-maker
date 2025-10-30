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

	mrn, ptName, order, err := c.FindPatientItemInInput(2, "order")
	if err != nil {
		return err
	}

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

	err := c.IgnoredOrders.Remove(order)
	if err != nil {
		return err
	}

	fmt.Printf("%s removed from Ignored Orders List\n", order)
	return nil
}

func homeCommandRemovePrepullOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing order argument")
	}

	order := strings.Join(c.lastInput[2:], " ")

	err := c.PrepullOrders.Remove(order)
	if err != nil {
		return err
	}

	fmt.Printf("%s removed from Prepull Orders List\n", order)
	return nil

}

func homeCommandRemovePtSupplied(c *config) error {
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments\nSyntax: remove ptSupplied [pt name] [order]")
	}

	mrn, ptName, order, err := c.FindPatientItemInInput(2, "order")
	if err != nil {
		return err
	}

	err = c.PtSupplyOrders.RemoveOrder(mrn, order)
	if err != nil {
		return err
	}

	fmt.Printf("%s for %s removed from Pt Supplied list\n", order, ptName)
	return nil
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
