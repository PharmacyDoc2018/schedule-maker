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

func homeCommandMarkPtSupplied(c *config) error {
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments\nsyntax: mark ptSupplied [pt name] [medication]")
	}

	mrn, ptName, medication, err := c.FindPatientItemInInput(2, "medication")
	if err != nil {
		return err
	}

	err = c.PtSupplyOrders.AddOrder(mrn, medication)
	if err != nil {
		return err
	}

	fmt.Printf("%s marked as Pt Supplied for %s\n", medication, ptName)
	return nil
}

func homeCommandMarkOrder(c *config) error {
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments\nsyntax: mark order [pt name] [order]")
	}

	mrn, _, order, err := c.FindPatientItemInInput(2, "order")
	if err != nil {
		return err
	}

	updatedPatient, err := func(mrn, order string) (Patient, error) {
		patient := c.PatientList[mrn]
		for key, val := range patient.Orders {
			if val == order {
				patient.Orders[key] = "'" + order
				return patient, nil
			}
		}
		return Patient{}, fmt.Errorf("error. order %s not found", order)
	}(mrn, order)
	if err != nil {
		return err
	}

	c.PatientList[mrn] = updatedPatient
	return nil

}

func patientCommandMarkOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing order argument")
	}
	order := strings.Join(c.lastInput[2:], " ")

	mrn := c.location.allNodes[c.location.currentNodeID].name

	updatedPatient, err := func(mrn, order string) (Patient, error) {
		patient := c.PatientList[mrn]
		for key, val := range patient.Orders {
			if val == order {
				patient.Orders[key] = "'" + order
				return patient, nil
			}
		}
		return Patient{}, fmt.Errorf("error. order %s not found", order)
	}(mrn, order)
	if err != nil {
		return err
	}

	c.PatientList[mrn] = updatedPatient

	return nil
}

func patientCommandMarkPtSupplied(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing order argument")
	}

	medication := strings.Join(c.lastInput[2:], " ")
	mrn := c.location.allNodes[c.location.currentNodeID].name

	err := c.PtSupplyOrders.AddOrder(mrn, medication)
	if err != nil {
		return err
	}

	return nil

}
