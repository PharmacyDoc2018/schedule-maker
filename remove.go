package main

import (
	"fmt"
	"strings"
)

func homeCommandRemoveOrder(c *config) error {
	//
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments.\nExpected format: remove order [pt name] [order name]")
	}

	ptName, err := func() (string, error) {
		i := 3
		for i < len(c.patientNameMap) {
			if _, ok := c.patientNameMap[strings.Join(c.lastInput[2:i], " ")]; ok {
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
	err = c.RemoveOrder(mrn, order)
	if err != nil {
		return err
	}

	fmt.Printf("removed order %s from %s\n", order, ptName)

	return nil
}
