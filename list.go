package main

import (
	"fmt"
	"sort"
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
