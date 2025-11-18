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
		if isSameDay(patientList.Date, list.Date) {
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
		if isSameDay(patientList.Date, list.Date) {
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
	if err != nil { //-- if no save data -> check for excel files
		fmt.Println("no save data found. looking for excel file...")
		patientLists, err := pullDataFromExcel(c)
		if err != nil {
			return err
		}

		c.PatientLists = patientLists

		isSet := func(c *config) bool {
			for i, ptList := range c.PatientLists.Slices {
				if isSameDay(ptList.Date, time.Now()) {
					c.PatientList = c.PatientLists.Slices[i]
					return true
				}
			}
			return false
		}(c)

		if !isSet {
			lastPatientList, _ := c.PatientLists.Last()
			c.PatientList = lastPatientList
		}

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

		fmt.Println("checking for excel data...")
		excelPatientLists, err := pullDataFromExcel(c)
		if err != nil {
			fmt.Printf("unable to find excel data:\n%s\n", err.Error())
		} else {
			for _, item := range excelPatientLists.Slices {
				isMissingList := func(c *config, item PatientList) bool { //-- only adds patient list from excel files if patient list for that day doesn't already exist
					for _, list := range c.PatientLists.Slices {
						if isSameDay(list.Date, item.Date) {
							return false
						}
					}
					return true
				}(c, item)

				if isMissingList {
					err := c.PatientLists.Add(item)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
			}
		}

		todayPtList, err := func(ptLists PatientLists) (PatientList, error) {
			for _, list := range ptLists.Slices {
				if isSameDay(list.Date, time.Now()) {
					return list, nil
				}
			}
			return PatientList{}, fmt.Errorf("no lists for today found")
		}(c.PatientLists)

		if err != nil { //-- if no patients for today found -> load last
			fmt.Println(err.Error())
			fmt.Print("loading most recent patient list...")
			lastPatientList, _ := c.PatientLists.Last()
			c.PatientList = lastPatientList
		} else {
			c.PatientList = todayPtList
		}

	}
	fmt.Println()
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
