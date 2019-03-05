package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"
)

const (
	clusterOperatorConditionChangeRegExp = "([A-Z][a-z]{2} [0-3][0-9] [012][0-9]:[0-5][0-9]:[0-5][0-9]).[0-9]{3} [EWI] clusteroperator/([a-z-]+) changed ([A-Za-z]+) to ([A-Za-z]+)"
)

func main() {
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	r := regexp.MustCompile(clusterOperatorConditionChangeRegExp)
	store := NewStore()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if r.MatchString(text) {
			match := r.FindStringSubmatch(text)
			timestamp, err := time.Parse("Jan 2 15:04:05 2006", fmt.Sprintf("%s %d", match[1], time.Now().Year()))
			if err != nil {
				fmt.Println(err)
				return
			}
			operator := match[2]
			conditionType := match[3]
			value := match[4]
			if value == "True" {
				store.Add(operator, conditionType, conditionType, text, timestamp)
			} else {
				store.Add(operator, conditionType, value, text, timestamp)
			}
		}
	}
	jsonData, err := store.toJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonFile, err := os.Create("data.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
}

type event struct {
	value       string
	description string
	timestamp   time.Time
}

type store struct {
	sync.Mutex
	events map[string]map[string][]event
}

func NewStore() *store {
	return &store{events: map[string]map[string][]event{}}
}

func (s *store) Add(group, label, value, description string, timestamp time.Time) {
	s.Lock()
	defer s.Unlock()
	groupevents, ok := s.events[group]
	if !ok {
		s.events[group] = map[string][]event{}
		groupevents, _ = s.events[group]
	}
	labelevents, ok := groupevents[label]
	if !ok || labelevents[len(labelevents)-1].value != value {
		event := event{value: value, timestamp: timestamp, description: description}
		groupevents[label] = append(groupevents[label], event)
	}
}

type LabelData struct {
	TimeRange [2]time.Time `json:"timeRange,omitempty"`
	Val       string       `json:"val,omitempty"`
	Extended  string       `json:"extended,omitempty"`
}

type GroupData struct {
	Label string      `json:"label,omitempty"`
	Data  []LabelData `json:"data,omitempty"`
}

type Group struct {
	Group string      `json:"group,omitempty"`
	Data  []GroupData `json:"data,omitempty"`
}

func (s *store) toJSON() ([]byte, error) {
	// find last event
	var endTime time.Time
	for _, labels := range s.events {
		for _, events := range labels {
			for _, event := range events {
				if event.timestamp.After(endTime) {
					endTime = event.timestamp
				}
			}
		}
	}
	endTime = endTime.Add(3 * time.Minute)

	var groups []Group
	for group, labels := range s.events {
		g := Group{Group: group}
		for label, events := range labels {
			gd := GroupData{Label: label}
			for i, event := range events {
				ld := LabelData{TimeRange: [2]time.Time{event.timestamp, endTime}, Val: event.value, Extended: event.description}
				gd.Data = append(gd.Data, ld)
				if i > 0 {
					gd.Data[i-1].TimeRange[1] = event.timestamp
				}
			}
			g.Data = append(g.Data, gd)
		}
		groups = append(groups, g)
	}
	return json.Marshal(groups)
}
