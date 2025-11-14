package main

import (
	"fmt"
	"strings"
	"time"
)

func homeCommandChangeApptTimeInf(c *config) error {
	if len(c.lastInput) < 4 {
		return fmt.Errorf("error. too few arguments.\n Syntax: change apptTimeInf [name] [new time]")
	}

	mrn, err := c.FindPatientInInput(2)
	if err != nil {
		return err
	}

	timePos := len(strings.Split(c.PatientList.Map[mrn].Name, " ")) + 2
	if len(c.lastInput) == timePos {
		return fmt.Errorf("error. no time entered.\n Syntax: change apptTimeInf [name] [new time]")
	}

	timeString := strings.Join(c.lastInput[timePos:], " ")
	newApptTimeOnly, err := time.Parse(timeFormat, timeString)
	if err != nil {
		return err
	}

	patient := c.PatientList.Map[mrn]
	for key, val := range patient.AppointmentTimes {
		if strings.Contains(key, infusionAppointmentTag) {
			oldApptTime := val
			newApptTime := time.Date(
				oldApptTime.Year(),
				oldApptTime.Month(),
				oldApptTime.Day(),
				newApptTimeOnly.Hour(),
				newApptTimeOnly.Minute(),
				0,
				0,
				oldApptTime.Location(),
			)

			if oldApptTime.Equal(newApptTime) {
				return fmt.Errorf("error. %s's infusion appointment is already scheduled at %s", patient.Name, newApptTime.Format(timeFormat))
			}

			for key := range patient.AppointmentTimes {
				if strings.Contains(key, infusionAppointmentTag) {
					patient.AppointmentTimes[key] = newApptTime
				}
			}

			c.PatientList.Map[mrn] = patient
			fmt.Printf("infusion appointment for %s changed from %s to %s.\n", patient.Name, oldApptTime.Format(timeFormat), newApptTime.Format(timeFormat))
			return nil
		}
	}

	return fmt.Errorf("error. infusion appointment not found for %s", patient.Name)
}

func patientCommandChangeApptTimeInf(c *config) error {
	if len(c.lastInput) < 3 {
		return fmt.Errorf("error. too few arguments. missing new appointment time")
	}

	timeString := strings.Join(c.lastInput[2:], " ")
	newApptTimeOnly, err := time.Parse(timeFormat, timeString)
	if err != nil {
		return err
	}

	mrn := c.location.allNodes[c.location.currentNodeID].name
	patient := c.PatientList.Map[mrn]

	for key, val := range patient.AppointmentTimes {
		if strings.Contains(key, infusionAppointmentTag) {
			oldApptTime := val
			newApptTime := time.Date(
				oldApptTime.Year(),
				oldApptTime.Month(),
				oldApptTime.Day(),
				newApptTimeOnly.Hour(),
				newApptTimeOnly.Minute(),
				0,
				0,
				oldApptTime.Location(),
			)

			if oldApptTime.Equal(newApptTime) {
				return fmt.Errorf("error. %s's infusion appointment is already scheduled at %s", patient.Name, newApptTime.Format(timeFormat))
			}

			for key := range patient.AppointmentTimes {
				if strings.Contains(key, infusionAppointmentTag) {
					patient.AppointmentTimes[key] = newApptTime
				}
			}

			c.PatientList.Map[mrn] = patient
			fmt.Printf("infusion appointment for %s changed from %s to %s.\n", patient.Name, oldApptTime.Format(timeFormat), newApptTime.Format(timeFormat))
			return nil

		}
	}

	return fmt.Errorf("error. infusion appointment not found for %s", patient.Name)
}
