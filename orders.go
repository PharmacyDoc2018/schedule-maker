package main

import (
	"math/rand"
	"strconv"
)

type missingOrdersQueue struct {
	queue []string
	len   int
}

func (m *missingOrdersQueue) AddPatient(mrn string) {
	m.queue = append(m.queue, mrn)
	m.len += 1
}

func (m *missingOrdersQueue) PopPatient() {
	m.queue = m.queue[1:]
	m.len -= 1

	if len(m.queue) == 0 {
		m.Clear()
	}
}

func (m *missingOrdersQueue) Clear() {
	m.queue = nil
}

func (m *missingOrdersQueue) NextPatient() string {
	return m.queue[0]
}

func (c *config) AddOrder(mrn, orderNum, orderName string) {
	c.patientList[mrn].orders[orderNum] = orderName
}

func (c *config) AddOrderQuick(mrn, orderName string) {
	//for adding orders without an order number
	randNum := rand.Intn(90000000) + 10000000
	pseudoOrderNum := "U" + strconv.Itoa(randNum)

	c.AddOrder(mrn, pseudoOrderNum, orderName)
}

func (c *config) FindMissingOrders() {
	const noOrders = 0
	for mrn := range c.patientList {
		if len(c.patientList[mrn].orders) == noOrders {
			c.missingOrders.AddPatient(mrn)
		}
	}
}
