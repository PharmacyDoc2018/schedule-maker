package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const infusionAppointmentTag = "AUBL INF"

type Patient struct {
	Mrn              string               `json:"mrn"`
	Name             string               `json:"name"`
	AppointmentTimes map[string]time.Time `json:"appointment_times"`
	Orders           map[string]string    `json:"orders"`
}

func initScheduledPatients(c *config) error {
	_, err := os.Stat(c.pathToSave)
	fmt.Println("looking for saved data...")
	if err == nil {
		fmt.Println("saved data found! pulling schedule...")
		data, err := os.ReadFile(c.pathToSave)
		if err != nil {
			return err
		}

		savedSchedule := map[string]Patient{}
		err = json.Unmarshal(data, &savedSchedule)
		if err != nil {
			return err
		}

		c.PatientList = savedSchedule

	} else if os.IsNotExist(err) {
		fmt.Println("no save data found. looking for excel file...")
		scheduleRows, ordersRows, err := pullDataFromExcel(c)
		if err != nil {
			return err
		}

		fmt.Println("creating patient list...")
		err = c.createPatientList(scheduleRows, ordersRows)
		if err != nil {
			return err
		}
	}

	fmt.Println("finding missing orders...")
	c.FindMissingInfusionOrders()
	fmt.Println("found", c.missingOrders.len, "patient(s) with missing orders...")

	return nil
}

func pullDataFromExcel(c *config) (scheduleRows, ordersRows [][]string, err error) {
	entries, err := os.ReadDir(c.pathToSch)
	if err != nil {
		return [][]string{}, [][]string{}, err
	}

	if len(entries) == 0 {
		return [][]string{}, [][]string{}, fmt.Errorf("no excel files found")
	}

	if len(entries) > 2 {
		return [][]string{}, [][]string{}, fmt.Errorf("too many files found in excel folder")
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

	fmt.Println("excel files found! pulling data...")

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
	if _, ok := c.PatientList[mrn]; ok {
		return fmt.Errorf("patient already exists")
	}

	appointmentTimeMap := make(map[string]time.Time)
	ordersMap := make(map[string]string)
	c.PatientList[mrn] = Patient{
		Mrn:              mrn,
		Name:             name,
		AppointmentTimes: appointmentTimeMap,
		Orders:           ordersMap,
	}

	return nil
}

func (c *config) addAppointment(mrn, schedule, date, time string) error {
	apptDateTime, err := parseDateTime(date, time)
	if err != nil {
		return err
	}

	c.PatientList[mrn].AppointmentTimes[schedule] = apptDateTime
	return nil
}

func (c *config) createPatientList(scheduleRows, ordersRows [][]string) error {
	// only called on init
	for _, row := range scheduleRows[1:] {
		if _, ok := c.PatientList[row[0]]; !ok {
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
		if _, ok := c.PatientList[row[3]]; !ok {
			c.createPatient(row[3], row[0])
		}

		c.AddOrder(row[3], row[7], row[6])
	}

	c.savePatientList()
	return nil
}

func (c *config) savePatientList() error {
	data, err := json.Marshal(c.PatientList)
	if err != nil {
		return err
	}

	saveFile, err := os.OpenFile(c.pathToSave, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() error {
		err = saveFile.Close()
		if err != nil {
			return err
		}
		return nil
	}()

	_, err = saveFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

type Schedule struct {
	table          [][]string
	colSpaceBuffer int
}

// row[0] == time
// row[1] == MRN
// row[2] == name
// row[4] == orders

func (s Schedule) LongestRowLen() int {
	longestSoFar := 0
	for _, row := range s.table {
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
	for _, row := range s.table {
		if len(row[2]) > longestSoFar {
			longestSoFar = len(row[2])
		}
	}

	return longestSoFar
}

func (s Schedule) LongestOrderName() int {
	longestSoFar := 0
	for _, row := range s.table {
		if len(row[3]) > longestSoFar {
			longestSoFar = len(row[3])
		}
	}

	return longestSoFar
}

func (s Schedule) Print() {
	timeColBuffer := 5 + s.colSpaceBuffer
	mrnColBuffer := 7 + s.colSpaceBuffer

	nameColBuffer := s.LongestPatientName() + s.colSpaceBuffer
	orderColBuffer := s.LongestOrderName() + s.colSpaceBuffer

	rowSeperator := " "
	rowSeperatorLen := timeColBuffer + mrnColBuffer + nameColBuffer + orderColBuffer + s.colSpaceBuffer + 1
	for i := 0; i < rowSeperatorLen; i++ {
		rowSeperator += "-"
	}

	cTop := 0
	cBottom := 0

	for cBottom < len(s.table)-1 {
		fmt.Println(rowSeperator)
		// -- set cTop and cBottom --
		row := s.table[cBottom]
		for s.table[cBottom+1][0] == "" {
			cBottom++
			if cBottom == len(s.table)-1 {
				break
			}
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
		timeColBufferSpaces := ""
		for i := 0; i < timeColBuffer; i++ {
			timeColBufferSpaces += " "
		}

		mrnColBufferSpaces := ""
		for i := 0; i < mrnColBuffer; i++ {
			mrnColBufferSpaces += " "
		}

		nameColBufferSpaces := ""
		for i := 0; i < nameColBuffer; i++ {
			nameColBufferSpaces += " "
		}

		printLine := int(((cBottom+1)-cTop)/2) + cTop
		for i := cTop; i <= cBottom; i++ {
			totalBuffer = orderColBuffer - len(s.table[i][3])
			frontBuffer = int(totalBuffer / 2)
			backBuffer = totalBuffer - frontBuffer

			orderColText := ""
			for j := 0; j < frontBuffer; j++ {
				orderColText += " "
			}

			orderColText += s.table[i][3]
			for j := 0; j < backBuffer; j++ {
				orderColText += " "
			}
			if i == printLine {
				fmt.Printf("|%s|%s|%s|%s|\n", timeColText, mrnColText, nameColText, orderColText)
			} else {
				fmt.Printf("|%s|%s|%s|%s|\n", timeColBufferSpaces, mrnColBufferSpaces, nameColBufferSpaces, orderColText)
			}
		}
		cBottom++
		cTop = cBottom
	}
	fmt.Println(rowSeperator)
}
