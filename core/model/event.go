package model

import (
	"fortuna/structure"
	"fmt"
)

type Identity struct {
	OrganizationID		string  `json:"organization_id"`
	WorkspaceID		string	`json:"workspace_id"`
	UserID			string	`json:"user_id"`
}

type EventSpec struct {
	InterfaceID	int			`json:"interface_id"`
	Version		string			`json:"version"`
	Params		*structure.OrderedMap	`json:"params"`
}

func NewEventSpecFromMap(data *structure.OrderedMap) (*EventSpec, error) {
	interface_id, ok := data.Get("interface_id")
	if !ok {
		return nil, fmt.Errorf("event spec must have interface_id")
	}

	interface_id, ok = interface_id.(int)
	if !ok {
		return nil, fmt.Errorf("interface_id must have type: integer")
	}

	version, ok := data.Get("version")
	if !ok {
		return nil, fmt.Errorf("event spec must have version")
	}

	params, ok := data.Get("params")
	if !ok {
		return nil, fmt.Errorf("event spec must have params")
	}

	params, ok = params.(*structure.OrderedMap)
	if !ok {
		return nil, fmt.Errorf("params must have type: ordered map")
	}

	spec := EventSpec{
		InterfaceID: interface_id.(int),
		Version: version.(string),
		Params: params.(*structure.OrderedMap),
	}

	return &spec, nil
}

type Event struct {
	Publisher 	Identity		`json:"publisher"`
	SpaceID		string			`json:"space_id"`

	Payload		*structure.OrderedMap	`json:"payload"`
	Spec		EventSpec		`json:"spec"`

	Topic		string			`json:"topic"`	
	SubTopic	string			`json:"subtopic"`
	Seperator	string			`json:"seperator"`
	Tag		string			`json:"tag"`
}

func (event *Event) Buffer() string {
	return ""
}

func NewEventFromOrderedMap(data *structure.OrderedMap) (*Event, error) {
	payload, ok := data.Get("payload")
	if !ok {
		return nil, fmt.Errorf("event data must have key:payload")
	}

	payload_s, ok := payload.(*structure.OrderedMap)
	if !ok {
		return nil, fmt.Errorf("payload must have type: string")
	}

	spaceID, ok := data.Get("space_id")
	if !ok {
		return nil, fmt.Errorf("event data must have key:space_id")
	}

	spaceID_s, ok := spaceID.(string)
	if !ok {
		return nil, fmt.Errorf("spaceID must have type: string")
	}

	_, ok = data.Get("publisher")
	if !ok {
		return nil, fmt.Errorf("event data must have key:publisher")
	}

	specData, ok := data.Get("spec")
	if !ok {
		return nil, fmt.Errorf("event data must have key:spec")
	}

	specData_s, ok := specData.(*structure.OrderedMap)

	topic, ok := data.Get("topic")
	if !ok {
		return nil, fmt.Errorf("event data must have key:topic")
	}

	topic_s, ok := topic.(string)
	if !ok {
		return nil, fmt.Errorf("topic must have type: string")
	}

	spec, err := NewEventSpecFromMap(specData_s)
	if err != nil {
		return nil, err
	}

	event := Event{
		SpaceID: spaceID_s,
		Spec: *spec,
		Payload: payload_s,
		Topic: topic_s,
	}

	return &event, nil
}

type EventExecutionError struct {
	Code		int
	Message 	string
}

type EventExecutionResult struct {
	Event		*Event
	Result		string
	Err		*EventExecutionError
}


