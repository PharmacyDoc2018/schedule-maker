package main

import (
	"fmt"
	"os"
	"path"
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

	fileStats, err := os.Stat(file.Path)
	if err != nil {
		return fmt.Errorf("error. cannot access doc stats for %s", file.Path)
	}

	fileName := fileStats.Name()

	if len(sheetList) > 1 {
		return fmt.Errorf("error. %s has many too many sheets", fileName)
	}

	if len(sheetList) == 0 {
		return fmt.Errorf("error. %s has no sheets. how did you manage that?", fileName)
	}

	sheet := sheetList[0]
	rows, err := file.GetRows(sheet)
	if err != nil {
		return err
	}

	isScheduleXLSX := func([][]string) bool {
		if rows[0][0] == "MRN" && rows[0][1] == "Patient" {
			return true
		}
		return false
	}(rows)

	isOrdersXLSX := func([][]string) bool {
		if rows[0][0] == "Patient" && rows[0][1] == "Age" {
			return true
		}
		return false
	}(rows)

	if isScheduleXLSX && isOrdersXLSX {
		return fmt.Errorf("error determining report type: %s", file.Path)
	}

	if !isScheduleXLSX && !isOrdersXLSX {
		return fmt.Errorf("error. unknown report type: %s", file.Path)
	}

	if isScheduleXLSX {
		visitDate, err := time.Parse("01/02/2006", rows[1][3])
		if err != nil {
			return err
		}

		scheduleDate := time.Date(
			visitDate.Year(),
			visitDate.Month(),
			visitDate.Day(),
			0,
			0,
			0,
			0,
			visitDate.Location(),
		)

		for i, entry := range e.Slices {
			if entry.Date.Year() == scheduleDate.Year() &&
				entry.Date.Month() == scheduleDate.Month() &&
				entry.Date.Day() == scheduleDate.Day() {
				if entry.Schedule == nil {
					e.Slices[i].Schedule = file
					e.Slices[i].Complete = true
					return nil
				}

				currentFileInfo, currentFileErr := os.Stat(file.Path)
				savedFileInfo, savedFilesErr := os.Stat(entry.Schedule.Path)
				if currentFileErr != nil || savedFilesErr != nil {
					return fmt.Errorf("error. schedule for that day already exists: cannot determine more recent file")
				}

				if savedFileInfo.ModTime().After(currentFileInfo.ModTime()) {
					return fmt.Errorf("error. current saved schedule from %s is more recent than schedule from %s", savedFileInfo.Name(), currentFileInfo.Name())

				} else {
					e.Slices[i].Schedule = file
					return nil
				}

			}
		}

		e.Slices = append(e.Slices, ExcelMatch{
			Date:     scheduleDate,
			Schedule: file,
			Complete: false,
		})

		return nil
	}

	if isOrdersXLSX {
		visitDate, err := time.Parse("01/02/2006", rows[1][10])
		if err != nil {
			return err
		}

		scheduleDate := time.Date(
			visitDate.Year(),
			visitDate.Month(),
			visitDate.Day(),
			0,
			0,
			0,
			0,
			visitDate.Location(),
		)

		for i, entry := range e.Slices {
			if entry.Date.Year() == scheduleDate.Year() &&
				entry.Date.Month() == scheduleDate.Month() &&
				entry.Date.Day() == scheduleDate.Day() {
				if entry.Orders == nil {
					e.Slices[i].Orders = file
					e.Slices[i].Complete = true
					return nil
				}

				currentFileInfo, currentFileErr := os.Stat(file.Path)
				savedFileInfo, savedFilesErr := os.Stat(entry.Schedule.Path)
				if currentFileErr != nil || savedFilesErr != nil {
					return fmt.Errorf("error. orders for that day already exists: cannot determine more recent file")
				}

				if savedFileInfo.ModTime().After(currentFileInfo.ModTime()) {
					return fmt.Errorf("error. current saved orders from %s is more recent than orders from %s", savedFileInfo.Name(), currentFileInfo.Name())

				} else {
					e.Slices[i].Orders = file
					return nil
				}

			}
		}

		e.Slices = append(e.Slices, ExcelMatch{
			Date:     scheduleDate,
			Schedule: file,
			Complete: false,
		})

		return nil
	}

	return fmt.Errorf("error. add entry failed: unknown error")

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

func pullDataFromExcel(c *config) (PatientLists, error) {
	entries, err := os.ReadDir(c.pathToSch)
	if err != nil {
		return PatientLists{}, err
	}

	if len(entries) == 0 {
		return PatientLists{}, fmt.Errorf("no excel files found")
	}

	fmt.Println("excel files found! pulling data...")

	files := []*excelize.File{}
	for _, entry := range entries {
		file, err := excelize.OpenFile(path.Join(c.pathToSch, entry.Name()))
		if err != nil {
			fmt.Println("unable to open %s", entry.Name())
			continue
		}
		files = append(files, file)
	}

	excelMatchList := ExcelMatchList{}
	errs := excelMatchList.AddEntries(files)
	for _, err := range errs {
		fmt.Println(err.Error())
	}

	fmt.Println("creating patient lists...")
	patientLists := PatientLists{}

	for _, item := range excelMatchList.Slices {
		if !item.Complete {
			continue
		}

		scheduleRows, err := item.Schedule.GetRows(item.Schedule.GetSheetList()[0])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		ordersRows, err := item.Orders.GetRows(item.Orders.GetSheetList()[0])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		ptList, err := createPatientList(scheduleRows, ordersRows)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		ptList.Date = item.Date

		patientLists.Add(ptList)
	}

	for _, file := range files {
		err := file.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return patientLists, nil

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
