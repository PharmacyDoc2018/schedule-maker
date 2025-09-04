package main

import (
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	lastInput   []string
	pathToSch   string
	patientList map[string]Patient
}

type commandMapList map[string]cliCommand

type Patient struct {
	mrn              string
	name             string
	appointmentTimes map[string]time.Time
	orders           []string
}
