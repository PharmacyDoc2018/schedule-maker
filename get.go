package main

import (
	"fmt"
	"sort"
	"strings"
)

func homeCommandGetScheduleInf(c *config) {
	schedule := c.CreateSchedule()

	schedule.colSpaceBuffer = 2

	filters := []string{"defaultOrderFilter", "defaultPatientFilterDone"}
	if len(c.lastInput) > 3 {
		args := c.lastInput[3:]

		for _, arg := range args {
			switch arg {
			case "--allOrders", "-ao":
				for i := range filters {
					if filters[i] == "defaultOrderFilter" {
						filters = append(filters[:i], filters[i+1:]...)
						break
					}
				}

			case "--allPatients", "-ap":
				for i := range filters {
					if filters[i] == "defaultPatientFilterDone" {
						filters = append(filters[:i], filters[i+1:]...)
						break
					}
				}

			default:
				fmt.Printf("error: unknown filter %s\n", arg)
				return
			}
		}

	}

	commandClear(c)
	schedule.Print(c, filters)
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

	pt := c.PatientList.Map[mrn].Name
	fmt.Printf("next patient with missing orders: %s (%s)\n", pt, mrn)

	return nil
}

func homeCommandGetPrepullOrders(c *config) error {
	type prePullLine struct {
		time      string
		visitType string
		name      string
		order     string
	}

	prePullList := []prePullLine{}

	for _, patient := range c.PatientList.Map {
		visitType := ""
		time := ""
		name := patient.Name
		for key, apptTime := range patient.AppointmentTimes {
			time = apptTime.Format(timeFormat)
			if strings.Contains(key, infusionAppointmentTag) {
				visitType = "INF"
				break
			} else {
				visitType = "CLINIC"
			}
		}

		for _, order := range patient.Orders {
			trimmedOrder := strings.ReplaceAll(strings.ToLower(order), " ", "")
			for _, prepullOrder := range c.PrepullOrders.List {
				if strings.Contains(trimmedOrder, prepullOrder) {
					printedOrder := order
					if c.PtSupplyOrders.IsPatientSupplied(patient.Mrn, order) {
						printedOrder += " ** Pt Supply **"
					}
					prePullList = append(prePullList, prePullLine{
						time:      time,
						visitType: visitType,
						name:      name,
						order:     printedOrder,
					})
				}
			}
		}
	}

	sort.Slice(prePullList, func(i, j int) bool {
		return prePullList[i].order < prePullList[j].order
	})

	timeBuffer := 8
	typeBuffer := 7
	nameBuffer := 0
	orderBuffer := 0

	for _, item := range prePullList {
		if len(item.name) > nameBuffer {
			nameBuffer = len(item.name)
		}

		if len(item.order) > orderBuffer {
			orderBuffer = len(item.order)
		}
	}
	nameBuffer += 4
	orderBuffer += 4

	for _, item := range prePullList {
		totalBuffer := timeBuffer - len(item.time)
		backBuffer := int(totalBuffer / 2)
		frontBuffer := totalBuffer - backBuffer
		timeText := ""
		for i := 0; i < frontBuffer; i++ {
			timeText += " "
		}
		timeText += item.time
		for i := 0; i < backBuffer; i++ {
			timeText += " "
		}

		totalBuffer = typeBuffer - len(item.visitType)
		backBuffer = int(totalBuffer / 2)
		frontBuffer = totalBuffer - backBuffer
		typeText := ""
		for i := 0; i < frontBuffer; i++ {
			typeText += " "
		}
		typeText += item.visitType
		for i := 0; i < backBuffer; i++ {
			typeText += " "
		}

		totalBuffer = nameBuffer - len(item.name)
		backBuffer = int(totalBuffer / 2)
		frontBuffer = totalBuffer - backBuffer
		nameText := ""
		for i := 0; i < frontBuffer; i++ {
			nameText += " "
		}
		nameText += item.name
		for i := 0; i < backBuffer; i++ {
			nameText += " "
		}

		totalBuffer = orderBuffer - len(item.order)
		backBuffer = int(totalBuffer / 2)
		frontBuffer = totalBuffer - backBuffer
		orderText := ""
		for i := 0; i < frontBuffer; i++ {
			orderText += " "
		}
		orderText += item.order
		for i := 0; i < backBuffer; i++ {
			orderText += " "
		}

		fmt.Printf("%s%s%s%s\n", timeText, typeText, nameText, orderText)
	}

	return nil
}

func homeCommandGetOrders(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. too few arguments\nExpected format: get orders [partial order name]")
	}

	orderSearchable := strings.ReplaceAll(strings.ToLower(strings.Join(c.lastInput[2:], " ")), " ", "")

	schedule := c.CreateSchedule()
	schedule.colSpaceBuffer = 2

	lastTime := ""
	lastMRN := ""
	lastName := ""

	newTable := [][]string{}

	for _, row := range schedule.table {
		if row[0] != "" {
			lastTime = row[0]
		}

		if row[1] != "" {
			lastMRN = row[1]
		}

		if row[2] != "" {
			lastName = row[2]
		}

		currentOrder := strings.ReplaceAll(strings.ToLower(row[3]), " ", "")
		if strings.Contains(currentOrder, orderSearchable) {
			newTable = append(newTable, []string{
				lastTime,
				lastMRN,
				lastName,
				row[3],
			})
		}
	}

	schedule.table = newTable
	schedule.Print(c, []string{})
	return nil
}
