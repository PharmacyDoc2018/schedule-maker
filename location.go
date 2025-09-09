package main

import (
	"math/rand"
)

const mainNodeID = 0

type LocationNode struct {
	id       int
	name     string
	patient  Patient
	parentID int
}

type Location struct {
	allNodes      map[int]*LocationNode
	currentNodeID int
}

func (l *Location) Path() string {
	path := ""
	node := l.allNodes[l.currentNodeID]
	for node.id != mainNodeID {
		path = node.name + " > " + path
		node = l.allNodes[node.parentID]
	}
	path = node.name + " > " + path

	return path
}

func (l *Location) NewNode(name string, parentID int) (newNodeID int) {
	newNodeID = rand.Intn(9000) + 1000

	l.allNodes[newNodeID] = &LocationNode{
		id:       newNodeID,
		name:     name,
		parentID: parentID,
	}

	return newNodeID
}

func (l *Location) NewPatientNode(c *config, mrn string, parentID int) (newNodeID int) {
	patient := c.patientList[mrn]

	newNodeID = l.NewNode(mrn, parentID)
	l.allNodes[newNodeID].patient = patient

	return newNodeID
}
