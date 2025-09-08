package main

type missingOrdersQueue struct {
	queue []string
	len   int
}

func (m *missingOrdersQueue) AddPatient(mrn string) {
	m.queue = append(m.queue, mrn)
	m.len += 1
}

func (m *missingOrdersQueue) PopPatient() {
	m.queue = m.queue[1:]
	m.len -= 1

	if len(m.queue) == 0 {
		m.Clear()
	}
}

func (m *missingOrdersQueue) Clear() {
	m.queue = nil
}
