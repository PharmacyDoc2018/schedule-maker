package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExcelMatch struct {
	Date     time.Time
	Schedule *excelize.File
	Orders   *excelize.File
	Complete bool
}

type ExcelMatchList struct {
	Slices []ExcelMatch
}

func (e *ExcelMatchList) addEntry(file *excelize.File) error {
	sheetList := file.GetSheetList()

	docProps, err := file.GetDocProps()
	if err != nil {
		return fmt.Errorf("error. cannot access doc properties for %s", file.Path)
	}

	if len(sheetList) > 1 {
		if err != nil {
			return fmt.Errorf("%s\nerror. workbook has too many sheets", err.Error())
		}
		return fmt.Errorf("error. %s has many too many sheets", docProps.Title)
	}

	if len(sheetList) == 0 {
		if err != nil {
			return fmt.Errorf("%s\nerror. workbook has no sheets. how did you manage that?", err.Error())
		}
		return fmt.Errorf("error. %s has no sheets. how did you manage that?", docProps.Title)
	}

	sheet := sheetList[0]
	rows, err := file.GetRows(sheet)
	if err != nil {
		return err
	}

	isScheduleXLSX, err := func([][]string) (bool, error) {
		//
	}(rows)

}

func (e *ExcelMatchList) AddEntries(files []*excelize.File) []error {
	errors := []error{}
	for _, file := range files {
		err := e.addEntry(file)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
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

func createPatientList(scheduleRows, ordersRows [][]string) (PatientList, error) {
	patientList := PatientList{}
	patientList.Map = map[string]Patient{}

	for _, row := range scheduleRows[1:] {
		if _, ok := patientList.Map[row[0]]; !ok {
			patientList.addPatient(row[0], row[1])
		}

		appointments := strings.Split(row[4], "\n")
		for _, appointment := range appointments {
			err := patientList.addAppointment(row[0], appointment, row[2], row[5])
			if err != nil {
				return PatientList{}, err
			}
		}
	}

	for _, row := range ordersRows[1:] {
		if _, ok := patientList.Map[row[3]]; !ok {
			patientList.addPatient(row[3], row[0])
		}

		if len(row[7]) != orderNumberLength {
			patientList.AddOrderQuick(row[3], row[6])
		} else {
			patientList.AddOrder(row[3], row[7], row[6])
		}
	}

	return patientList, nil
}
