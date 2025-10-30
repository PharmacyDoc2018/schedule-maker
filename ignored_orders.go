package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type IgnoredOrders struct {
	Map map[string]struct{} `json:"ignored_orders_list"`
}

type OldIgnoredOrders struct {
	List []string `json:"ignored_orders_list"`
}

func (i *IgnoredOrders) Add(order string) error {
	trimmedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	if _, ok := i.Map[trimmedOrder]; ok {
		return fmt.Errorf("error. %s already in Ignored Orders list", order)
	}

	i.Map[trimmedOrder] = struct{}{}
	return nil
}

func (i *IgnoredOrders) Remove(order string) error {
	trimmedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	if _, ok := i.Map[trimmedOrder]; !ok {
		return fmt.Errorf("error. %s not found in Ignored Orders list", order)
	}

	delete(i.Map, trimmedOrder)
	return nil
}

func (i *IgnoredOrders) Len() int {
	return len(i.Map)
}

func (i *IgnoredOrders) Slices() []string {
	orderSlice := []string{}
	for order := range i.Map {
		orderSlice = append(orderSlice, order)
	}

	return orderSlice
}

func (i *IgnoredOrders) Exists(order string) bool {
	trimmedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	_, ok := i.Map[trimmedOrder]
	return ok
}

func (c *config) PullIgnoredOrdersList() error {
	_, err := os.Stat(c.pathToIgnoredOrders)
	if err == nil {
		data, err := os.ReadFile(c.pathToIgnoredOrders)
		if err != nil {
			return err
		}

		ignoredOrders := IgnoredOrders{}
		err = json.Unmarshal(data, &ignoredOrders)
		if err != nil {
			oldIgnoredOrders := OldIgnoredOrders{}
			ignoredOrders.Map = map[string]struct{}{}
			err = json.Unmarshal(data, &oldIgnoredOrders)
			if err != nil {
				return err
			}

			for _, order := range oldIgnoredOrders.List {
				ignoredOrders.Map[order] = struct{}{}
			}

		}

		c.IgnoredOrders = ignoredOrders

	} else {
		return fmt.Errorf("warning: ignored orders list not found")
	}

	return nil
}

func (c *config) saveIgnoredOrdersList() error {
	data, err := json.Marshal(c.IgnoredOrders)
	if err != nil {
		return err
	}

	saveFile, err := os.OpenFile(c.pathToIgnoredOrders, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
