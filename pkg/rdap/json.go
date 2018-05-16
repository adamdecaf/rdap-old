package rdap

import (
	"time"
)

type LinkJSON struct {
	Value string `json:"value"`
	Rel   string `json:"rel"`
	Href  string `json:"href"`
	Type  string `json:"type"`
}

type RemarkJSON struct {
	Description []string `json:"description,omitempty"`
}

type EventJSON struct {
	EventAction string    `json:"eventAction,omitempty"`
	EventDate   time.Time `json:"eventDate,omitempty"`
}

type EntityJSON struct {
	ObjectClassName string        `json:"objectClassName"`
	Handle          string        `json:"handle"`
	VcardArray      []interface{} `json:"vcardArray"`
	Roles           []string      `json:"roles"`
	Remarks         []struct {
		Description []string `json:"description"`
	} `json:"remarks"`
	Links  []LinkJSON `json:"links"`
	Events []struct {
		EventAction string    `json:"eventAction"`
		EventDate   time.Time `json:"eventDate"`
	} `json:"events"`
}

type IPNetworkJSON struct {
	ObjectClassName string       `json:"objectClassName"`
	Handle          string       `json:"handle,omitempty"`
	StartAddress    string       `json:"startAddress,omitempty"`
	EndAddress      string       `json:"endAddress,omitempty"`
	IPVersion       string       `json:"ipVersion,omitempty"`
	Name            string       `json:"name,omitempty"`
	Type            string       `json:"type,omitempty"`
	Country         string       `json:"country,omitempty"`
	ParentHandle    string       `json:"parentHandle,omitempty"`
	Status          []string     `json:"status,omitempty"`
	Remarks         []RemarkJSON `json:"remarks,omitempty"`
	Links           []LinkJSON   `json:"links,omitempty"`
	Events          []EventJSON  `json:"events,omitempty"`
	Entities        []EntityJSON `json:"entities,omitempty"`
}
