package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PtSupplyOrders struct {
	Map map[string]map[string]struct{} `json:"map"`
}

func (p *PtSupplyOrders) IsPatientSupplied(mrn, order string) bool {
	trimmedOrder := strings.ToLower(order)

	ptSuppliedOrders, ok := p.Map[mrn]
	if !ok {
		return ok
	}

	for ptSuppliedOrder := range ptSuppliedOrders {
		if strings.Contains(trimmedOrder, ptSuppliedOrder) {
			return true
		}
	}

	return false
}

func (p *PtSupplyOrders) AddOrder(mrn, medication string) error {
	trimmedMed := strings.ToLower(medication)

	if _, ok := p.Map[mrn]; !ok {
		p.Map[mrn] = map[string]struct{}{}
	} else {
		for med := range p.Map[mrn] {
			if med == trimmedMed {
				return fmt.Errorf("error. medication already marked as patient supplied")
			}
		}
	}

	p.Map[mrn][trimmedMed] = struct{}{}
	return nil
}

func (p *PtSupplyOrders) RemoveOrder(mrn, order string) error {
	trimmedOrder := strings.ToLower(order)

	if _, ok := p.Map[mrn]; !ok {
		return fmt.Errorf("error. patient does not have any patient supplied medications")
	}

	for med := range p.Map[mrn] {
		if strings.Contains(trimmedOrder, med) {
			delete(p.Map[mrn], med)
			if len(p.Map[mrn]) == 0 {
				delete(p.Map, mrn)
			}
			return nil
		}
	}

	return fmt.Errorf("error. patient supplied %s not found for that patient", order)
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
