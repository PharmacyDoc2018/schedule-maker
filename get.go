package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func homeCommandGetScheduleInf(c *config) {
	schedule := Schedule{
		colSpaceBuffer: 2,
	}

	type infAppt struct {
		time   string
		mrn    string
		name   string
		orders []string
	}

	infApptSlices := []infAppt{}
	for _, patient := range c.PatientList {
		for appt, apptTime := range patient.AppointmentTimes {
			if strings.Contains(appt, infusionAppointmentTag) {
				ordersSlice := []string{}
				for _, order := range patient.Orders {
					ordersSlice = append(ordersSlice, order)
				}
				infApptSlices = append(infApptSlices, infAppt{
					time:   apptTime.Format(timeFormat),
					mrn:    patient.Mrn,
					name:   patient.Name,
					orders: ordersSlice,
				})
				break
			}
		}
	}

	sort.Slice(infApptSlices, func(i, j int) bool {
		a, _ := time.Parse(timeFormat, infApptSlices[i].time)
		b, _ := time.Parse(timeFormat, infApptSlices[j].time)
		return a.Before(b)
	})

	for _, appt := range infApptSlices {
		if len(appt.orders) > 0 {
			schedule.table = append(schedule.table, []string{
				appt.time,
				appt.mrn,
				appt.name,
				appt.orders[0],
			})
			for _, order := range appt.orders[1:] {
				schedule.table = append(schedule.table, []string{
					"",
					"",
					"",
					order,
				})
			}
		} else {
			schedule.table = append(schedule.table, []string{
				appt.time,
				appt.mrn,
				appt.name,
				"",
			})
		}

	}

	if len(c.lastInput) > 3 {
		thirdArg := c.lastInput[3]

		switch thirdArg {
		case "allOrders", "-ao":
			commandClear(c)
			schedule.Print(c, []string{})
			return

		default:
			fmt.Printf("error: unknown filter %s\n", thirdArg)
			return
		}
	}

	commandClear(c)
	schedule.Print(c, []string{"default"})
}

func homeCommandGetNextMissingOrderPatient(c *config) error {
	mrn, err := c.missingOrders.NextPatient()
	if err != nil {
		return err
	}

	err = c.location.SelectPatientNode(mrn)
	if err != nil {
		return err
	}

	pt := c.PatientList[mrn].Name
	fmt.Printf("next patient with missing orders: %s (%s)\n", pt, mrn)

	return nil
}
