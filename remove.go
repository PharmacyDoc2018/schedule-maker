package main

import (
	"fmt"
	"os"
	"strings"
	"time"
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

	err = c.PatientList.RemoveOrder(mrn, order)
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

func homeCommandRemovePatientList(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing date argument\nExpected format: remove ptList [date]")
	}

	ptListDateString := c.lastInput[2]
	ptListDate, err := time.Parse(dateFormat, ptListDateString)
	if err != nil {
		return err
	}

	listToRemove := PatientList{}
	listToRemove.Date = ptListDate
	for _, ptList := range c.PatientLists.Slices {
		if isSameDay(ptList.Date, listToRemove.Date) {
			err := c.PatientLists.RemoveList(listToRemove)
			if err != nil {
				return err
			}
			fmt.Printf("patient list for %s removed\n", ptListDateString)
			commandSave(c)
			return nil
		}
	}

	return fmt.Errorf("error. no patient list found for %s", ptListDateString)
}

func homeCommandRemovePatient(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing patient name")
	}

	mrn, err := c.FindPatientInInput(2)
	if err != nil {
		return err
	}

	ptName := c.PatientList.Map[mrn].Name

	err = c.PatientList.removePatient(mrn)
	if err != nil {
		return err
	}

	fmt.Printf("%s has been removed from the current patient list\n", ptName)
	return nil
}

func homeCommandRemoveDone(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. missing patient name")
	}

	mrn, err := c.FindPatientInInput(2)
	if err != nil {
		return err
	}

	if !c.PatientList.Map[mrn].VisitComplete {
		return fmt.Errorf("error. day for %s not completed", c.PatientList.Map[mrn].Name)
	}

	patient := c.PatientList.Map[mrn]
	patient.VisitComplete = false
	c.PatientList.Map[mrn] = patient

	fmt.Printf("%s added back to the schedule\n", patient.Name)

	return nil
}

func patientCommandRemoveOrder(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. too few arguments.\nExpected format: remove order [pt name] [order name]")
	}

	order := strings.Join(c.lastInput[2:], " ")
	mrn := c.location.allNodes[c.location.currentNodeID].name
	patient := c.PatientList.Map[mrn]

	for key, val := range patient.Orders {
		if val == order {
			delete(patient.Orders, key)
			c.PatientList.Map[mrn] = patient
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

func patientCommandRemoveDone(c *config) error {
	mrn := c.location.allNodes[c.location.currentNodeID].name

	if !c.PatientList.Map[mrn].VisitComplete {
		return fmt.Errorf("error. day for %s not completed", c.PatientList.Map[mrn].Name)
	}

	patient := c.PatientList.Map[mrn]
	patient.VisitComplete = false
	c.PatientList.Map[mrn] = patient

	return nil
}
