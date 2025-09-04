package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func pullData(c *config) error {
	entries, err := os.ReadDir(c.pathToSch)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no excel files found")
	}

	if len(entries) > 2 {
		return fmt.Errorf("too many files found")
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
	fmt.Println("schedulePath: ", schedulePath) // -- DELETE
	fmt.Println("ordersPath: ", ordersPath)     // -- DELETE

	if schedulePath == "" {
		return fmt.Errorf("no schedule excel file found")
	}

	if ordersPath == "" {
		return fmt.Errorf("no orders excel file found")
	}

	scheduleXLSX, err := excelize.OpenFile(schedulePath)
	if err != nil {
		return err
	}
	defer func() {
		err := scheduleXLSX.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	scheduleSheetList := scheduleXLSX.GetSheetList()
	if len(scheduleSheetList) > 1 {
		return fmt.Errorf("too many sheets in schedule workbook")
	}
	if len(scheduleSheetList) == 0 {
		return fmt.Errorf("no sheets in schedule workbook. how did you manage that?")
	}

	scheduleSheet := scheduleSheetList[0]
	scheduleRows, err := scheduleXLSX.GetRows(scheduleSheet)
	if err != nil {
		return err
	}
	c.scheduleRows = scheduleRows

	ordersXLSX, err := excelize.OpenFile(ordersPath)
	if err != nil {
		return err
	}
	defer func() {
		err := ordersXLSX.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	ordersSheetList := ordersXLSX.GetSheetList()
	if len(ordersSheetList) > 1 {
		return fmt.Errorf("too many sheets in orders workbook")
	}
	if len(ordersSheetList) == 0 {
		return fmt.Errorf("no sheets in orders workbook. how did you manage that?")
	}

	ordersSheet := ordersSheetList[0]
	ordersRows, err := ordersXLSX.GetRows(ordersSheet)
	if err != nil {
		return err
	}
	c.ordersRows = ordersRows

	return nil
}
