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
