package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const noOrders = 0

type missingOrdersQueue struct {
	queue []string
}

func (m *missingOrdersQueue) Len() int {
	return len(m.queue)
}

func (m *missingOrdersQueue) AddPatient(mrn string) {
	m.queue = append(m.queue, mrn)
}

func (m *missingOrdersQueue) PopPatient() error {
	if m.queue == nil {
		return fmt.Errorf("no more missing orders")
	}
	m.queue = m.queue[1:]

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
			if m.Len() == 0 {
				m.Clear()
			}
			return nil
		}
	}

	return fmt.Errorf("patient not found in missing order queue")

}

func (m *missingOrdersQueue) Clear() {
	m.queue = nil
}

func (m *missingOrdersQueue) Sort(c *config, key, order string) error {
	switch key {
	case "time", "appointmentTime":
		switch order {
		case "asc", "ascending":
			sort.Slice(m.queue, func(i, j int) bool {
				aTime, _ := func() (time.Time, error) {
					for appt, apptTime := range c.PatientList[m.queue[i]].AppointmentTimes {
						if strings.Contains(appt, infusionAppointmentTag) {
							return apptTime, nil
						}
					}
					return time.Time{}, fmt.Errorf("not found")
				}()

				bTime, _ := func() (time.Time, error) {
					for appt, apptTime := range c.PatientList[m.queue[j]].AppointmentTimes {
						if strings.Contains(appt, infusionAppointmentTag) {
							return apptTime, nil
						}
					}
					return time.Time{}, fmt.Errorf("not found")
				}()

				return aTime.Before(bTime)
			})
			return nil

		default:
			return fmt.Errorf("order not found for sorting by infusion appointment time")
		}

	default:
		return fmt.Errorf("sort key not valid")
	}
}

func (m *missingOrdersQueue) NextPatient() (string, error) {
	if m.queue == nil {
		return "", fmt.Errorf("no patients left in queue")
	}
	return m.queue[0], nil
}

func (c *config) AddOrder(mrn, orderNum, orderName string) {
	c.PatientList[mrn].Orders[orderNum] = orderName
}

func (c *config) AddOrderQuick(mrn, orderName string) {
	//for adding orders without an order number
	randNum := rand.Intn(90000000) + 10000000
	pseudoOrderNum := "U" + strconv.Itoa(randNum)

	c.AddOrder(mrn, pseudoOrderNum, orderName)
}

func (c *config) FindMissingOrders() {
	for mrn := range c.PatientList {
		if len(c.PatientList[mrn].Orders) == noOrders {
			c.missingOrders.AddPatient(mrn)
		}
	}
}

func (c *config) FindMissingInfusionOrders() {
	for mrn := range c.PatientList {
		if len(c.PatientList[mrn].Orders) == noOrders {
			for appt := range c.PatientList[mrn].AppointmentTimes {
				if strings.Contains(appt, infusionAppointmentTag) {
					c.missingOrders.AddPatient(mrn)
					break
				}
			}
		}
	}
	c.missingOrders.Sort(c, "time", "asc")
}

func (c *config) PullIgnoredOrdersList() error {
	_, err := os.Stat(c.pathToIgnoredOrders)
	if err == nil {
		data, err := os.ReadFile(c.pathToSave)
		if err != nil {
			return err
		}

		ignoredOrders := []string{}
		err = json.Unmarshal(data, &ignoredOrders)
		if err != nil {
			return err
		}

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

	saveFile, err := os.OpenFile(c.pathToSave, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
