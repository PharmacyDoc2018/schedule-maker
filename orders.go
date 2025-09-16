package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const noOrders = 0

type missingOrdersQueue struct {
	queue []string
	len   int
}

func (m *missingOrdersQueue) AddPatient(mrn string) {
	m.queue = append(m.queue, mrn)
	m.len += 1
}

func (m *missingOrdersQueue) PopPatient() error {
	if m.queue == nil {
		return fmt.Errorf("no more missing orders")
	}
	m.queue = m.queue[1:]
	m.len -= 1

	if len(m.queue) == 0 {
		m.Clear()
	}

	return nil
}

func (m *missingOrdersQueue) RemovePatient(mrn string) error {
	if m.queue == nil {
		return fmt.Errorf("no more missing orders")
	}

	for i := range m.queue {
		if m.queue[i] == mrn {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("patient not found in missing order queue")

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
	for mrn := range c.patientList {
		if len(c.patientList[mrn].orders) == noOrders {
			c.missingOrders.AddPatient(mrn)
		}
	}
}

func (c *config) FindMissingInfusionOrders() {
	for mrn := range c.patientList {
		if len(c.patientList[mrn].orders) == noOrders {
			for appt := range c.patientList[mrn].appointmentTimes {
				if strings.Contains(appt, infusionAppointmentTag) {
					c.missingOrders.AddPatient(mrn)
				}
			}
		}

	}
}
