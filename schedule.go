package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func initScheduledPatients(c *config) error {
	fmt.Println("pulling data from excel files...")
	scheduleRows, ordersRows, err := pullData(c)
	if err != nil {
		return err
	}

	fmt.Println("creating patient list...")
	err = createPatientList(c, scheduleRows, ordersRows)
	if err != nil {
		return err
	}

	for key, val := range c.patientList["1003917"].appointmentTimes {
		fmt.Println(key, val)
	}
	fmt.Println("MRN: ", c.patientList["1003917"].mrn)
	fmt.Println("Name: ", c.patientList["1003917"].name)

	return nil
}

func pullData(c *config) (scheduleRows, ordersRows [][]string, err error) {
	entries, err := os.ReadDir(c.pathToSch)
	if err != nil {
		return [][]string{}, [][]string{}, err
	}

	if len(entries) == 0 {
		return [][]string{}, [][]string{}, fmt.Errorf("no excel files found")
	}

	if len(entries) > 2 {
		return [][]string{}, [][]string{}, fmt.Errorf("too many files found")
	}

	var schedulePath string
	var ordersPath string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "Schedule__Augusta") {
			schedulePath = filepath.Join(c.pathToSch, entry.Name())
		}
		if strings.Contains(entry.Name(), "Scheduled_Orders__Augusta") {
			ordersPath = filepath.Join(c.pathToSch, entry.Name())
		}
	}

	if schedulePath == "" {
		return [][]string{}, [][]string{}, fmt.Errorf("no schedule excel file found")
	}

	if ordersPath == "" {
		return [][]string{}, [][]string{}, fmt.Errorf("no orders excel file found")
	}

	scheduleXLSX, err := excelize.OpenFile(schedulePath)
	if err != nil {
		return [][]string{}, [][]string{}, err
	}
	defer func() {
		err := scheduleXLSX.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	scheduleSheetList := scheduleXLSX.GetSheetList()
	if len(scheduleSheetList) > 1 {
		return [][]string{}, [][]string{}, fmt.Errorf("too many sheets in schedule workbook")
	}
	if len(scheduleSheetList) == 0 {
		return [][]string{}, [][]string{}, fmt.Errorf("no sheets in schedule workbook. how did you manage that?")
	}

	scheduleSheet := scheduleSheetList[0]
	scheduleRows, err = scheduleXLSX.GetRows(scheduleSheet)
	if err != nil {
		return [][]string{}, [][]string{}, err
	}

	ordersXLSX, err := excelize.OpenFile(ordersPath)
	if err != nil {
		return [][]string{}, [][]string{}, err
	}
	defer func() {
		err := ordersXLSX.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	ordersSheetList := ordersXLSX.GetSheetList()
	if len(ordersSheetList) > 1 {
		return [][]string{}, [][]string{}, fmt.Errorf("too many sheets in orders workbook")
	}
	if len(ordersSheetList) == 0 {
		return [][]string{}, [][]string{}, fmt.Errorf("no sheets in orders workbook. how did you manage that?")
	}

	ordersSheet := ordersSheetList[0]
	ordersRows, err = ordersXLSX.GetRows(ordersSheet)
	if err != nil {
		return [][]string{}, [][]string{}, err
	}

	return scheduleRows, ordersRows, nil
}

func createPatientList(c *config, scheduleRows, ordersRows [][]string) error {
	const timeFormat = "3:04 PM"
	for _, row := range scheduleRows[1:] {
		if _, ok := c.patientList[row[0]]; !ok {
			apptTime, err := time.Parse(timeFormat, row[5])
			if err != nil {
				return err
			}
			c.patientList[row[0]] = Patient{
				mrn:  row[0],
				name: row[1],
				appointmentTimes: map[string]time.Time{
					row[4]: apptTime,
				},
				orders: []string{},
			}
		} else {
			apptTime, err := time.Parse(timeFormat, row[5])
			if err != nil {
				return err
			}
			c.patientList[row[0]].appointmentTimes[row[4]] = apptTime
		}
	}

	return nil
}
