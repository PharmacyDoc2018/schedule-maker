package main

import (
	"fmt"
	"sort"
)

func homeCommandListIgnoredOrders(c *config) error {
	if len(c.IgnoredOrders.List) == 0 {
		return fmt.Errorf("error. No ignored orders")
	}

	commandClear(c)
	fmt.Println("Ignored Orders:")
	fmt.Println()

	ignoredOrders := c.IgnoredOrders.List
	sort.Slice(ignoredOrders, func(i int, j int) bool {
		return ignoredOrders[i] < ignoredOrders[j]
	})

	for _, order := range ignoredOrders {
		fmt.Println(order)
	}

	return nil
}
