package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Patient struct {
	Mrn              string               `json:"mrn"`
	Name             string               `json:"name"`
	AppointmentTimes map[string]time.Time `json:"appointment_times"`
	Orders           map[string]string    `json:"orders"`
	VisitComplete    bool                 `json:"visit_complete"`
}

type PatientList struct {
	Map  map[string]Patient `json:"map"`
	Date time.Time          `json:"date"`
}

func getPatientListFileNameFromDate(date time.Time) string {
	return strconv.Itoa(date.Year()) + strconv.Itoa(int(date.Month())) + strconv.Itoa(date.Day()) + ".json"

}

func (p *PatientList) addPatient(mrn, name string) error {
	if _, ok := p.Map[mrn]; ok {
		return fmt.Errorf("patient already exists")
	}

	appointmentTimeMap := make(map[string]time.Time)
	ordersMap := make(map[string]string)
	p.Map[mrn] = Patient{
		Mrn:              mrn,
		Name:             name,
		AppointmentTimes: appointmentTimeMap,
		Orders:           ordersMap,
	}

	return nil
}

func (p *PatientList) addAppointment(mrn, schedule, date, time string) error {
	apptDateTime, err := parseDateTime(date, time)
	if err != nil {
		return err
	}

	p.Map[mrn].AppointmentTimes[schedule] = apptDateTime
	return nil
}

func (p *PatientList) AddOrder(mrn, orderNum, orderName string) {
	p.Map[mrn].Orders[orderNum] = orderName
}

func (p *PatientList) AddOrderQuick(mrn, orderName string) {
	//for adding orders without an order number
	randNum := rand.Intn(90000000) + 10000000
	pseudoOrderNum := "U" + strconv.Itoa(randNum)

	p.AddOrder(mrn, pseudoOrderNum, orderName)
}

func (p *PatientList) RemoveOrder(mrn, orderName string) error {
	for key, val := range p.Map[mrn].Orders {
		if val == orderName {
			delete(p.Map[mrn].Orders, key)
			return nil
		}
	}
	return fmt.Errorf("order %s not found for %s", orderName, p.Map[mrn].Name)

}

func (p *PatientList) FindMissingOrders() missingOrdersQueue {
	moq := missingOrdersQueue{}
	for mrn := range p.Map {
		if len(p.Map[mrn].Orders) == noOrders {
			moq.AddPatient(mrn)
		}
	}

	return moq
}

func (p *PatientList) FindMissingInfusionOrders() missingOrdersQueue {
	moq := missingOrdersQueue{}
	for mrn := range p.Map {
		if len(p.Map[mrn].Orders) == noOrders {
			for appt := range p.Map[mrn].AppointmentTimes {
				if strings.Contains(appt, infusionAppointmentTag) {
					moq.AddPatient(mrn)
					break
				}
			}
		}
	}
	moq.Sort(p, "time", "asc")
	return moq
}

type PatientLists struct {
	Slices []PatientList `json:"slices"`
}

func (p *PatientLists) Add(patientList PatientList) error {
	for _, list := range p.Slices {
		if patientList.Date.Year() == list.Date.Year() &&
			patientList.Date.Month() == list.Date.Month() &&
			patientList.Date.Day() == list.Date.Day() {
			return fmt.Errorf("error. patient list already exists for this day")
		}
	}
	p.Slices = append(p.Slices, patientList)
	return nil
}

func (p *PatientLists) Last() (PatientList, error) {
	if len(p.Slices) == 0 {
		return PatientList{}, fmt.Errorf("error. no patient lists found")
	}

	return p.Slices[len(p.Slices)-1], nil
}

func (p *PatientLists) Update(patientList PatientList) error {
	for i, list := range p.Slices {
		if patientList.Date.Year() == list.Date.Year() &&
			patientList.Date.Month() == list.Date.Month() &&
			patientList.Date.Day() == list.Date.Day() {
			p.Slices = append(p.Slices[:i], p.Slices[i+1:]...)
			p.Slices = append(p.Slices, patientList)
			return nil
		}
	}

	return fmt.Errorf("error. patient list for %s not found", patientList.Date.Format(dateFormat))

}

func initPatientLists(c *config) error {
	_, err := os.Stat(c.pathToSave)
	fmt.Println("looking for saved data...")
	if err != nil {
		fmt.Println("no save data found. looking for excel file...")
		scheduleRows, ordersRows, err := pullDataFromExcel(c)
		if err != nil {
			return err //-- no save data, no excel files breakpoint
		}

		fmt.Println("creating patient list...")
		patientList, err := createPatientList(scheduleRows, ordersRows)
		if err != nil {
			return err //-- no dave data, error with excel files breakpoint
		}

		randApptTime := time.Time{}
		for _, pt := range patientList.Map {
			for _, apptTime := range pt.AppointmentTimes {
				randApptTime = apptTime
				break
			}
			break
		}
		dateOnly := time.Date(
			randApptTime.Year(),
			randApptTime.Month(),
			randApptTime.Day(),
			0,
			0,
			0,
			0,
			randApptTime.Location(),
		)

		patientList.Date = dateOnly
		c.PatientList = patientList
		c.PatientLists = PatientLists{}
		c.PatientLists.Add(patientList)

	} else { //-- if save data is found
		fmt.Println("saved data found! Loading patient lists...")

		patientLists := PatientLists{}
		data, err := os.ReadFile(c.pathToSave) //-- read save data
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &patientLists) //-- unmarshal data
		if err != nil {
			return err
		}

		//-- OLD SCHEDULE CLEANUP

		c.PatientLists = patientLists //-- set unmarshalled patient lists into config

		todayPtList, err := func(ptLists PatientLists) (PatientList, error) {
			for _, list := range ptLists.Slices {
				if list.Date.Year() == time.Now().Year() &&
					list.Date.Month() == time.Now().Month() &&
					list.Date.Day() == time.Now().Day() {
					return list, nil
				}
			}
			return PatientList{}, fmt.Errorf("no lists for today found")
		}(c.PatientLists)

		if err != nil { //-- no patient list for today, check excel folder
			fmt.Println(err.Error())
			fmt.Println("checking for excel files...")

			scheduleRows, ordersRows, err := pullDataFromExcel(c)
			if err != nil { //-- no patient list for today, no excel files
				fmt.Println(err.Error())
				fmt.Print("loading most recent patient list...")
				c.PatientList, err = c.PatientLists.Last()
				if err != nil {
					return err
				}
			}

			fmt.Println("creating patient list...")
			patientList, err := createPatientList(scheduleRows, ordersRows)
			if err != nil { //-- no patient list for today, error with excel files
				fmt.Println(err.Error())
				fmt.Print("loading most recent patient list...")
				c.PatientList, err = c.PatientLists.Last()
				if err != nil {
					return err
				}
			}

			randApptTime := time.Time{}
			for _, pt := range patientList.Map {
				for _, apptTime := range pt.AppointmentTimes {
					randApptTime = apptTime
					break
				}
				break
			}
			dateOnly := time.Date(
				randApptTime.Year(),
				randApptTime.Month(),
				randApptTime.Day(),
				0,
				0,
				0,
				0,
				randApptTime.Location(),
			)

			patientList.Date = dateOnly
			c.PatientList = patientList
			c.PatientLists.Add(patientList)

		} else { //-- set list with today's date as active patient list
			c.PatientList = todayPtList
		}
	}
	return nil
}

func (c *config) createPatientNameMap() {
	for _, val := range c.PatientList.Map {
		c.patientNameMap[val.Name] = struct{}{}
	}
}

func (c *config) savePatientLists() error {
	c.PatientLists.Update(c.PatientList)
	data, err := json.Marshal(c.PatientLists)
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
