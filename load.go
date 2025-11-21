package main

import (
	"fmt"
	"time"
)

func homeCommandLoadExcelData(c *config) error {
	patientLists, err := pullDataFromExcel(c)
	if err != nil {
		return err
	}

	for i, ptList := range patientLists.Slices {
		if isSameDay(ptList.Date, time.Now()) {
			fmt.Println("error. excel data contains patient list for today. list will not be loaded")
			continue
		}

		isMissingList, foundPosition := func(c *config, ptList PatientList) (bool, int) { //-- only adds patient list from excel files if patient list for that day doesn't already exist
			for j, list := range c.PatientLists.Slices {
				if isSameDay(list.Date, ptList.Date) {
					return false, j
				}
			}
			return true, 0
		}(c, ptList)

		if isMissingList {
			err := c.PatientLists.Add(ptList)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			newPtList := patientLists.Slices[i]
			oldPtList := c.PatientLists.Slices[foundPosition]

			for mrn, patient := range newPtList.Map {
				if len(patient.Orders) == 0 { //-- if new list has no orders
					for orderNum, orderName := range oldPtList.Map[mrn].Orders {
						newPtList.Map[mrn].Orders[orderNum] = orderName
					}
				} else if len(patient.Orders) != len(oldPtList.Map[mrn].Orders) {
					for orderNum, orderName := range oldPtList.Map[mrn].Orders {
						runes := []rune(orderNum)
						if runes[0] == 'U' {
							newPtList.Map[mrn].Orders[orderNum] = orderName
						}
					}
				}
			}

			err := c.PatientLists.RemoveList(oldPtList)
			if err != nil {
				fmt.Println(err.Error())
			}

			err = c.PatientLists.Add(newPtList)
			if err != nil {
				fmt.Println(err.Error())
			}

		}
	}

	return nil
}
