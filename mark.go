package main

import (
	"fmt"
	"strings"
)

func homeCommandMarkDone(c *config) error {
	// changes VisitComplete bool for specified patient to true

	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. please enter patient name mark as done")
	}

	ptName := strings.Join(c.lastInput[2:], " ")
	var mrn string
	var patient Patient
	for key, pt := range c.PatientList {
		if pt.Name == ptName {
			mrn = key
			patient = pt
			break
		}
	}
	if mrn == "" {
		return fmt.Errorf("error. %s not found", ptName)
	}

	if c.PatientList[mrn].VisitComplete {
		return fmt.Errorf("day already completed for %s", ptName)
	}

	patient.VisitComplete = true
	c.PatientList[mrn] = patient
	fmt.Printf("day for %s has been completed\n", ptName)

	return nil

}
