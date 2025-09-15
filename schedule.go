package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type Patient struct {
	mrn              string
	name             string
	appointmentTimes map[string]time.Time
	orders           map[string]string
}

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
	c.FindMissingOrders()
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

		c.AddOrder(row[3], row[7], row[6])
	}

	return nil
}

type Schedule [][]string

// row[0] == time
// row[1] == MRN
// row[2] == name
// row[4] == orders

func (s Schedule) LongestRowLen() int {
	longestSoFar := 0
	for _, row := range s {
		rowLen := 0
		for _, item := range row {
			rowLen += len(item)
		}
		if rowLen > longestSoFar {
			longestSoFar = rowLen
		}
	}
	return longestSoFar
}

func (s Schedule) LongestPatientName() int {
	longestSoFar := 0
	for _, row := range s {
		if len(row[2]) > longestSoFar {
			longestSoFar = len(row[2])
		}
	}

	return longestSoFar
}

func (s Schedule) LongestOrderName() int {
	longestSoFar := 0
	for _, row := range s {
		if len(row[2]) > longestSoFar {
			longestSoFar = len(row[4])
		}
	}

	return longestSoFar
}

func (s Schedule) Print() {
	const timeColBuffer = 9
	const mrnColBuffer = 11

	nameColBuffer := s.LongestPatientName() + 4
	orderColBuffer := s.LongestOrderName() + 4

	rowSeperator := ""
	for i := 0; i < s.LongestRowLen(); i++ {
		rowSeperator += "_"
	}

	cTop := 0
	cBottom := 0

	fmt.Println(rowSeperator)
	for cBottom <= len(s)-1 {
		// -- set cTop and cBottom --
		row := s[cBottom]
		for s[cBottom+1][0] == "" {
			cBottom++
		}

		// -- Time Column Formatting --
		totalBuffer := timeColBuffer - len(row[0])
		backBuffer := int(totalBuffer / 2)
		frontBuffer := totalBuffer - backBuffer
		timeColText := ""
		for i := 0; i < frontBuffer; i++ {
			timeColText += " "
		}
		timeColText += row[0]
		for i := 0; i < backBuffer; i++ {
			timeColText += " "
		}

		// -- MRN Column Formatting --
		totalBuffer = mrnColBuffer - len(row[1])
		backBuffer = int(totalBuffer / 2)
		frontBuffer = totalBuffer - backBuffer
		mrnColText := ""
		for i := 0; i < frontBuffer; i++ {
			mrnColText += " "
		}
		mrnColText += row[1]
		for i := 0; i < backBuffer; i++ {
			mrnColText += " "
		}

		// -- Name Column Formatting --
		totalBuffer = nameColBuffer - len(row[2])
		frontBuffer = int(totalBuffer / 2)
		backBuffer = totalBuffer - frontBuffer
		nameColText := ""
		for i := 0; i < frontBuffer; i++ {
			nameColText += " "
		}
		nameColText += row[2]
		for i := 0; i < backBuffer; i++ {
			nameColText += " "
		}

		// -- Order Formating and Printing --
		printLine := int(((cBottom+1)-cTop)/2) + cBottom
		for i := cTop; i <= cBottom; i++ {
			totalBuffer = orderColBuffer - len(s[i][3])
			frontBuffer = int(totalBuffer / 2)
			backBuffer = totalBuffer + frontBuffer
			orderColText := "" // -- PICK UP HERE
			for j := 0; j < frontBuffer; j++ {
				//
			}
			if i == printLine {
				fmt.Printf("|%s%s%s%s|\n", timeColText, mrnColText, nameColText)
			}
		}
	}
}
