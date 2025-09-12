package main

import (
	"fmt"
	"math/rand"
)

const invalidNodeID = -1
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

func (l *Location) ChangeNodeLoc(name string) error {
	desiredLocationID := invalidNodeID

	for _, val := range l.allNodes {
		if val.name == name {
			desiredLocationID = val.id
		}
	}

	if desiredLocationID == invalidNodeID {
		return fmt.Errorf("invalid location")
	}

	for l.currentNodeID != desiredLocationID {
		oldCurrentNodeID := l.currentNodeID
		newCurrentNodeID := l.allNodes[l.currentNodeID].parentID
		l.currentNodeID = newCurrentNodeID
		err := l.CloseNode(oldCurrentNodeID)
		if err != nil {
			l.currentNodeID = oldCurrentNodeID
			return err
		}
	}

	return nil
}

func (l *Location) CloseNode(nodeID int) error {
	if nodeID == mainNodeID {
		return fmt.Errorf("cannot close home node")
	}

	if _, ok := l.allNodes[nodeID]; !ok {
		return fmt.Errorf("error. node not found")
	}

	if l.currentNodeID == nodeID {
		return fmt.Errorf("error. cannot close current location node")
	}

	delete(l.allNodes, nodeID)
	return nil
}

func (l *Location) SelectPatientNode(mrn string) error {
	if l.currentNodeID != mainNodeID {
		return fmt.Errorf("error. patient must be selected from home")
	}

	newNodeID := l.NewNode(mrn, l.currentNodeID)
	l.allNodes[newNodeID].locType = PatientLoc
	l.currentNodeID = newNodeID

	return nil
}
