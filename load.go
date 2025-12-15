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
				} else {
					manualOrderList := map[string]string{}
					markedOrderList := map[string]string{}
					for orderNum, orderName := range oldPtList.Map[mrn].Orders {
						runes := []rune(orderNum)
						orderNameRunes := []rune(orderName)
						if runes[0] == 'U' {
							manualOrderList[orderNum] = orderName
						} else if orderNameRunes[0] == '\'' {
							markedOrderList[orderNum] = orderName
						}
					}

					if len(manualOrderList) != 0 {
						for manOrderNum, manOrderName := range manualOrderList {
							needToAdd := func(manOrderName string) bool {
								for _, currentOrderName := range patient.Orders {
									if currentOrderName == manOrderName {
										return false
									}
								}
								return true
							}(manOrderName)
							if needToAdd {
								newPtList.Map[mrn].Orders[manOrderNum] = manOrderName
							}
						}
					}

					if len(markedOrderList) != 0 {
						for markedOrderNum, markedOrderName := range markedOrderList {
							if _, ok := patient.Orders[markedOrderNum]; ok {
								newPtList.Map[mrn].Orders[markedOrderNum] = markedOrderName
							}
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
