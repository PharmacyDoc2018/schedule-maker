package main

import (
	"fmt"
	"sort"
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

func (m *missingOrdersQueue) Sort(p *PatientList, key, order string) error {
	switch key {
	case "time", "appointmentTime":
		switch order {
		case "asc", "ascending":
			sort.Slice(m.queue, func(i, j int) bool {
				aTime, _ := func() (time.Time, error) {
					for appt, apptTime := range p.Map[m.queue[i]].AppointmentTimes {
						if strings.Contains(appt, infusionAppointmentTag) {
							return apptTime, nil
						} else if strings.Contains(appt, nurseAppointmentTag) {
							return apptTime, nil
						}
					}
					return time.Time{}, fmt.Errorf("not found")
				}()

				bTime, _ := func() (time.Time, error) {
					for appt, apptTime := range p.Map[m.queue[j]].AppointmentTimes {
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

func (c *config) OrderPreprocessing(order string) string {
	orderRunes := []rune(order)
	firstChar := string(orderRunes[:1])
	orderNoFirstChar := string(orderRunes[1:])

	switch firstChar {
	case ".":
		switch orderNoFirstChar {
		case "p":
			return "Phlebotomy therapeutic"

		case "u":
			return "unhook"

		case "f":
			return "flush"

		default:
			return order
		}

	default:
		return order
	}
}
