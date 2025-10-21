package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PtSupplyOrder struct {
	Mrn        string `json:"mrn"`
	Medication string `json:"medication"`
}

type PtSupplyOrders struct {
	List []PtSupplyOrder `json:"list"`
}

func (p *PtSupplyOrders) IsPatientSupplied(mrn, order string) bool {
	trimmedOrder := strings.ToLower(order)
	for _, order := range p.List {
		if order.Mrn == mrn {
			if strings.Contains(trimmedOrder, order.Medication) {
				return true
			}
		}
	}
	return false
}

func (p *PtSupplyOrders) AddOrder(mrn, medication string) error {
	trimmedMed := strings.ToLower(medication)

	newEntry := PtSupplyOrder{
		Mrn:        mrn,
		Medication: trimmedMed,
	}

	for _, order := range p.List {
		if order == newEntry {
			return fmt.Errorf("error. Pt Supplied Order already exists")
		}
	}

	p.List = append(p.List, newEntry)
	return nil
}

func (p *PtSupplyOrders) RemoveOrder(mrn, order string) error {
	trimmedOrder := strings.ToLower(order)

	for i, order := range p.List {
		if mrn == order.Mrn {
			if strings.Contains(trimmedOrder, order.Medication) {
				p.List = append(p.List[:i], p.List[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("error. Pt Supplied Order not found")
}

func (c *config) PullPtSupplyOrdersList() error {
	_, err := os.Stat(c.pathToPtSupplyOrders)
	if err == nil {
		data, err := os.ReadFile(c.pathToPtSupplyOrders)
		if err != nil {
			return err
		}

		ptSupplyOrders := PtSupplyOrders{}
		err = json.Unmarshal(data, &ptSupplyOrders)
		if err != nil {
			return err
		}

		c.PtSupplyOrders = ptSupplyOrders
	} else {
		return fmt.Errorf("warning: Pt Supply list not found")
	}
	return nil
}

func (c *config) savePtSupplyOrderList() error {
	data, err := json.Marshal(c.PtSupplyOrders)
	if err != nil {
		return err
	}

	saveFile, err := os.OpenFile(c.pathToPtSupplyOrders, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
