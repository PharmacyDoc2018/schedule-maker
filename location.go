package main

import (
	"math/rand"
)

const mainNodeID = 0

type LocationType int

const (
	Home LocationType = iota
	PatientLoc
	UnknownLoc
)

type LocationNode struct {
	id       int
	name     string
	locType  LocationType
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
		locType:  UnknownLoc,
		parentID: parentID,
	}

	return newNodeID
}

func (l *Location) NewPatientNode(mrn string, parentID int) (newNodeID int) {
	newNodeID = l.NewNode(mrn, parentID)
	l.allNodes[newNodeID].locType = PatientLoc

	return newNodeID
}
