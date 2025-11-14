package main

import (
	"fmt"
	"strings"
)

func commandSelectPatient(c *config) error {
	var pt string
	if len(c.lastInput) == 3 {
		pt = c.lastInput[2]
	} else {
		pt = strings.Join(c.lastInput[2:], " ")
	}

	if _, ok := c.PatientList.Map[pt]; ok {
		err := c.location.SelectPatientNode(pt)
		if err != nil {
			return err
		}
		return nil

	} else {
		for key, val := range c.PatientList.Map {
			if pt == val.Name {
				fmt.Println("Found patient with name. MRN is", key)
				fmt.Println()
				err := c.location.SelectPatientNode(key)
				if err != nil {
					return err
				}

			}
		}
		return nil
	}
}
