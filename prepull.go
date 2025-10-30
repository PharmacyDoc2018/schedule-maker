package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PrepullOrders struct {
	List []string `json:"prepull_orders_list"`
}

func (p *PrepullOrders) Add(order string) error {
	trimmedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	for _, item := range p.List {
		if trimmedOrder == item {
			return fmt.Errorf("entry already exists on the Prepull Orders list")
		}
	}

	p.List = append(p.List, trimmedOrder)
	return nil

}

func (p *PrepullOrders) Remove(order string) error {
	trimmedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")

	for i, item := range p.List {
		if strings.Contains(trimmedOrder, item) {
			p.List = append(p.List[:i], p.List[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("error. %s not found on the Prepull Orders List", order)

}

func (c *config) PullPrepullOrdersList() error {
	_, err := os.Stat(c.pathToPrepullOrders)
	if err == nil {
		data, err := os.ReadFile(c.pathToPrepullOrders)
		if err != nil {
			return err
		}

		prepullOrders := PrepullOrders{}
		err = json.Unmarshal(data, &prepullOrders)
		if err != nil {
			return err
		}

		c.PrepullOrders = prepullOrders

	} else {
		return fmt.Errorf("warning: prepull orders list not found")
	}

	return nil
}

func (c *config) savePrepullOrdersList() error {
	data, err := json.Marshal(c.PrepullOrders)
	if err != nil {
		return err
	}

	saveFile, err := os.OpenFile(c.pathToPrepullOrders, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
