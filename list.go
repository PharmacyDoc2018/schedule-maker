package main

import (
	"fmt"
	"sort"
	"time"
)

func homeCommandListIgnoredOrders(c *config) error {
	if c.IgnoredOrders.Len() == 0 {
		return fmt.Errorf("error. No ignored orders")
	}

	commandClear(c)
	fmt.Println("Ignored Orders:")
	fmt.Println()

	ignoredOrders := c.IgnoredOrders.Slices()

	sort.Slice(ignoredOrders, func(i int, j int) bool {
		return ignoredOrders[i] < ignoredOrders[j]
	})

	for _, order := range ignoredOrders {
		fmt.Println(order)
	}

	return nil
}

func homeCommandListPrepullOrders(c *config) error {
	if c.PrepullOrders.Len() == 0 {
		return fmt.Errorf("error. No prepull orders")
	}

	commandClear(c)
	fmt.Println("Prepull Orders:")
	fmt.Println()

	prepullOrders := c.PrepullOrders.ListOrders()

	sort.Slice(prepullOrders, func(i, j int) bool {
		return prepullOrders[i] < prepullOrders[j]
	})

	for _, order := range prepullOrders {
		fmt.Println(order)
	}

	return nil
}

func homeCommandListPatientLists(c *config) error {
	dates := c.PatientLists.GetDates()
	if len(dates) == 0 {
		return fmt.Errorf("error. no patient lists found")
	}

	sort.Slice(dates, func(i, j int) bool {
		dateStringOne := dates[i]
		dateStringTwo := dates[j]

		dateOne, err := time.Parse(dateFormat, dateStringOne)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		dateTwo, err := time.Parse(dateFormat, dateStringTwo)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		return dateOne.Before(dateTwo)
	})

	for _, date := range dates {
		fmt.Println(date)
	}
	return nil
}
