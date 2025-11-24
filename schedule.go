package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const infusionAppointmentTag = "AUBL INF"
const orderNumberLength = 9
const timeFormat = "3:04 PM"
const dateFormat = "01-02-06"

type Schedule struct {
	table          [][]string
	colSpaceBuffer int
}

const ptSupplyBuffer = 14

// row[0] == time
// row[1] == MRN
// row[2] == name
// row[3] == orders

func (s Schedule) LongestRowLen() int {
	longestSoFar := 0
	for _, row := range s.table {
		rowLen := 0
		for _, item := range row {
			rowLen += len(item)
		}
		if rowLen > longestSoFar {
			longestSoFar = rowLen
		}
	}
	return longestSoFar
}

func (s Schedule) LongestPatientName() int {
	longestSoFar := 0
	for _, row := range s.table {
		if len(row[2]) > longestSoFar {
			longestSoFar = len(row[2])
		}
	}

	return longestSoFar
}

func (s Schedule) LongestOrderName() int {
	longestSoFar := 0
	for _, row := range s.table {
		if len(row[3]) > longestSoFar {
			longestSoFar = len(row[3])
		}
	}

	return longestSoFar
}

func (s Schedule) OrderPrintFormat(order string) string {
	order = strings.ReplaceAll(order, "'", " ")

	if strings.Contains(order, "*") {
		starPos := 0
		for i, r := range order {
			if r == '*' {
				starPos = i
				break
			}
		}

		runes := []rune(order)
		lastRune := 0
		for i := len(runes) - 1; i > 0; i-- {
			if runes[i] != ' ' {
				lastRune = i
				break
			}
		}

		totalBuffer := starPos + (len(runes) - lastRune)
		newBuffer := totalBuffer - ptSupplyBuffer

		frontBuffer := int(newBuffer / 2)
		backBuffer := newBuffer - frontBuffer

		newOrder := ""
		for i := 0; i < frontBuffer; i++ {
			newOrder += " "
		}
		newOrder += "**Pt Supply** "
		newOrder += string(runes[starPos+1 : lastRune+1])
		for i := 0; i < backBuffer; i++ {
			newOrder += " "
		}

		order = newOrder
	}

	return order
}

func (s Schedule) Print(c *config, filters []string) {
	for _, filter := range filters {
		switch filter {
		case "defaultOrderFilter":
			newTable := [][]string{}
			i := 0
			for i < len(s.table) {
				row := s.table[i]
				if len(row) < 4 {
					newTable = append(newTable, row)
					i++
					continue
				}

				if c.IgnoredOrders.Exists(row[3]) {
					if row[0] != "" && i+1 < len(s.table) {
						nextRow := s.table[i+1]
						if nextRow[0] == "" {
							s.table[i+1][0] = row[0]
							s.table[i+1][1] = row[1]
							s.table[i+1][2] = row[2]
							i++
							continue
						} else if nextRow[0] != "" {
							s.table[i][3] = ""
							newTable = append(newTable, s.table[i])
						}
					}
				} else {
					newTable = append(newTable, s.table[i])
				}
				i++
			}

			s.table = newTable
		case "defaultPatientFilterDone":
			newTable := [][]string{}
			top := 0
			bottom := 0
			for top < len(s.table) {
				bottom = top
				for bottom+1 < len(s.table) && s.table[bottom+1][0] == "" {
					bottom++
				}

				mrn := s.table[top][1]
				if !c.PatientList.Map[mrn].VisitComplete {
					newTable = append(newTable, s.table[top:bottom+1]...)
				}

				top = bottom + 1
			}

			s.table = newTable
		}
	}

	timeColBuffer := 8 + s.colSpaceBuffer
	mrnColBuffer := 7 + s.colSpaceBuffer

	nameColBuffer := s.LongestPatientName() + s.colSpaceBuffer
	orderColBuffer := s.LongestOrderName() + s.colSpaceBuffer

	needsPtSuppliedBuffer := false
	currentMRN := ""
	for i, row := range s.table {
		if row[1] != "" {
			currentMRN = row[1]
		}
		if c.PtSupplyOrders.IsPatientSupplied(currentMRN, row[3]) {
			if s.LongestOrderName()-len(row[3]) < 14 {
				needsPtSuppliedBuffer = true
			}
			s.table[i][3] = "*" + row[3]
		}
	}

	if needsPtSuppliedBuffer {
		orderColBuffer += ptSupplyBuffer
	}

	rowSeperator := " "
	rowSeperatorLen := timeColBuffer + mrnColBuffer + nameColBuffer + orderColBuffer + s.colSpaceBuffer + 1
	for i := 0; i < rowSeperatorLen; i++ {
		rowSeperator += "-"
	}

	cTop := 0
	cBottom := 0

	for cTop < len(s.table) {
		cBottom = cTop
		for cBottom+1 < len(s.table) && s.table[cBottom+1][0] == "" {
			cBottom++
		}
		fmt.Println(rowSeperator)

		row := s.table[cTop]

		// -- Time Column Formatting --
		totalBuffer := timeColBuffer - len(row[0])
		backBuffer := int(totalBuffer / 2)
		frontBuffer := totalBuffer - backBuffer
		timeColText := ""
		for i := 0; i < frontBuffer; i++ {
			timeColText += " "
		}
		timeColText += row[0]
		for i := 0; i < backBuffer; i++ {
			timeColText += " "
		}

		// -- MRN Column Formatting --
		totalBuffer = mrnColBuffer - len(row[1])
		backBuffer = int(totalBuffer / 2)
		frontBuffer = totalBuffer - backBuffer
		mrnColText := ""
		for i := 0; i < frontBuffer; i++ {
			mrnColText += " "
		}
		mrnColText += row[1]
		for i := 0; i < backBuffer; i++ {
			mrnColText += " "
		}

		// -- Name Column Formatting --
		totalBuffer = nameColBuffer - len(row[2])
		frontBuffer = int(totalBuffer / 2)
		backBuffer = totalBuffer - frontBuffer
		nameColText := ""
		for i := 0; i < frontBuffer; i++ {
			nameColText += " "
		}
		nameColText += row[2]
		for i := 0; i < backBuffer; i++ {
			nameColText += " "
		}

		// -- Order Formating and Printing --
		timeColBufferSpaces := ""
		for i := 0; i < timeColBuffer; i++ {
			timeColBufferSpaces += " "
		}

		mrnColBufferSpaces := ""
		for i := 0; i < mrnColBuffer; i++ {
			mrnColBufferSpaces += " "
		}

		nameColBufferSpaces := ""
		for i := 0; i < nameColBuffer; i++ {
			nameColBufferSpaces += " "
		}

		printLine := int(((cBottom+1)-cTop)/2) + cTop
		for i := cTop; i <= cBottom; i++ {
			totalBuffer = orderColBuffer - len(s.table[i][3])
			frontBuffer = int(totalBuffer / 2)
			backBuffer = totalBuffer - frontBuffer

			orderColText := ""
			for j := 0; j < frontBuffer; j++ {
				orderColText += " "
			}

			orderColText += s.table[i][3]
			for j := 0; j < backBuffer; j++ {
				orderColText += " "
			}
			if i == printLine {
				fmt.Printf("|%s|%s|%s|%s|\n", timeColText, mrnColText, nameColText, s.OrderPrintFormat(orderColText))
			} else {
				fmt.Printf("|%s|%s|%s|%s|\n", timeColBufferSpaces, mrnColBufferSpaces, nameColBufferSpaces, s.OrderPrintFormat(orderColText))
			}
		}
		//cBottom++
		cTop = cBottom + 1
	}
	fmt.Println(rowSeperator)
}

func (c *config) CreateSchedule() Schedule {
	schedule := Schedule{}

	type infAppt struct {
		time   string
		mrn    string
		name   string
		orders []string
	}

	infApptSlices := []infAppt{}
	for _, patient := range c.PatientList.Map {
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

	return schedule
}
