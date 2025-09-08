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
	err = c.createPatientList(scheduleRows, ordersRows)
	if err != nil {
		return err
	}

	fmt.Println("finding missing orders...")
	c.findMissingOrders()
	fmt.Println("found", c.missingOrders.len, "patient(s) with missing orders...")

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

func parseDateTime(apptDateString, apptTimeString string) (time.Time, error) {
	const timeFormat = "3:04 PM"
	const dateFormat = "01-02-06"

	apptTime, err := time.Parse(timeFormat, apptTimeString)
	if err != nil {
		return time.Time{}, err
	}

	apptDate, err := time.Parse(dateFormat, apptDateString)
	if err != nil {
		return time.Time{}, err
	}

	apptDateTime := time.Date(
		apptDate.Year(),
		apptDate.Month(),
		apptDate.Day(),
		apptTime.Hour(),
		apptTime.Minute(),
		0,
		0,
		apptDate.Location(),
	)

	return apptDateTime, nil
}

func (c *config) createPatient(mrn, name string) error {
	if _, ok := c.patientList[mrn]; ok {
		return fmt.Errorf("patient already exists")
	}

	appointmentTimeMap := make(map[string]time.Time)
	ordersMap := make(map[string]string)
	c.patientList[mrn] = Patient{
		mrn:              mrn,
		name:             name,
		appointmentTimes: appointmentTimeMap,
		orders:           ordersMap,
	}

	return nil
}

func (c *config) addAppointment(mrn, schedule, date, time string) error {
	apptDateTime, err := parseDateTime(date, time)
	if err != nil {
		return err
	}

	c.patientList[mrn].appointmentTimes[schedule] = apptDateTime
	return nil
}

func (c *config) addOrder(mrn, orderNumber, orderName string) {
	c.patientList[mrn].orders[orderNumber] = orderName
}

func (c *config) createPatientList(scheduleRows, ordersRows [][]string) error {
	for _, row := range scheduleRows[1:] {
		if _, ok := c.patientList[row[0]]; !ok {
			c.createPatient(row[0], row[1])
		}

		appointments := strings.Split(row[4], "\n")
		for _, appointment := range appointments {
			err := c.addAppointment(row[0], appointment, row[2], row[5])
			if err != nil {
				return err
			}
		}

	}

	for _, row := range ordersRows[1:] {
		if _, ok := c.patientList[row[3]]; !ok {
			c.createPatient(row[3], row[0])
		}

		c.addOrder(row[3], row[7], row[6])
	}

	return nil
}

func (c *config) findMissingOrders() {
	const noOrders = 0
	for mrn := range c.patientList {
		if len(c.patientList[mrn].orders) == noOrders {
			c.missingOrders.AddPatient(mrn)
		}
	}
}
