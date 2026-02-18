package main

import (
	"fmt"
	"sort"
	"strings"
)

func homeCommandGetScheduleInf(c *config) error {
	schedule := c.CreateSchedule(c.PatientList)

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
				return fmt.Errorf("error: unknown filter %s", arg)
			}
		}

	}

	commandClear(c)
	schedule.Print(c, filters)
	return nil
}

func homeCommandGetScheduleNurse(c *config) error {
	schedule := c.CreateRNSchedule(c.PatientList)

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
				return fmt.Errorf("error: unknown filter %s", arg)
			}
		}

	}

	commandClear(c)
	schedule.Print(c, filters)
	return nil
}

func homeCommandGetScheduleClinic(c *config) error {
	schedule := c.CreateClinicSchedule(c.PatientList)

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
				return fmt.Errorf("error: unknown filter %s", arg)
			}
		}

	}

	commandClear(c)
	schedule.Print(c, filters)
	return nil
}

func homeCommandGetScheduleProvider(c *config, name string) error {
	schedule := c.CreateProviderSchedule(c.PatientList, name)

	schedule.colSpaceBuffer = 2

	filters := []string{"defaultOrderFilter", "defaultPatientFilterDone"}
	if len(c.lastInput) > 2 {
		args := c.lastInput[2:]

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

			}
		}

	}

	commandClear(c)
	schedule.Print(c, filters)
	return nil
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
			} else if strings.Contains(key, nurseAppointmentTag) {
				visitType = "NURSE"
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

	ptLists := []PatientList{}

	parsedInput := []string{}
	for _, word := range c.lastInput {
		if word == "-a" {
			ptLists = c.PatientLists.Slices
			continue
		}
		parsedInput = append(parsedInput, word)
	}

	c.lastInput = parsedInput

	if len(ptLists) == 0 {
		ptLists = append(ptLists, c.PatientList)
	} else {
		sort.Slice(ptLists, func(i, j int) bool {
			return ptLists[i].Date.Before(ptLists[j].Date)
		})
	}

	orderSearchable := strings.ReplaceAll(strings.ToLower(strings.Join(c.lastInput[2:], " ")), " ", "")

	for _, list := range ptLists {
		fmt.Printf("%s:\n", list.Date.Format(dateFormat))
		func(ptList PatientList) {
			schedule := c.CreateSchedule(ptList)
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

		}(list)
	}

	return nil
}

func homeCommandGetPtSupplied(c *config) error {
	if len(c.lastInput) > 3 {
		return fmt.Errorf("error. too many arguments")
	}

	ptLists := []PatientList{}

	if len(c.lastInput) == 3 {
		switch c.lastInput[2] {
		case "-a":
			ptLists = c.PatientLists.Slices
			sort.Slice(ptLists, func(i, j int) bool {
				return ptLists[i].Date.Before(ptLists[j].Date)
			})

		default:
			return fmt.Errorf("error. invalid argument %s", c.lastInput[2])
		}
	}

	if len(c.lastInput) == 2 {
		ptLists = []PatientList{c.PatientList}
	}

	for _, list := range ptLists {
		fmt.Printf("%s:\n", list.Date.Format(dateFormat))
		func(ptList PatientList) {
			schedule := c.CreateSchedule(ptList)
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

				if c.PtSupplyOrders.IsPatientSupplied(lastMRN, row[3]) {
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

		}(list)
	}

	return nil

}
