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
	InterfaceID	string			`json:"interface_id"`
	KernelVersion	string 			`json:"kernel_version"`
	Params		*structure.OrderedMap	`json:"params"`
}

func NewEventSpec(interface_id string, kernel_version string, params *structure.OrderedMap) *EventSpec {
	return &EventSpec{
		InterfaceID: interface_id,
		KernelVersion: kernel_version,
		Params: params,
	}
}

func (es *EventSpec) Buffer() []byte {
	return []byte(es.String())
}

func (es *EventSpec) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()

	om.Set("interface_id", es.InterfaceID)
	om.Set("version", es.KernelVersion)
	om.Set("params", es.Params)
	
	return om
}

func (es *EventSpec) String() string {
	om := es.Map()
	return om.Ser()
}

func NewEventSpecFromMap(data *structure.OrderedMap) (*EventSpec, error) {
	interface_id, ok := data.Get("interface_id")
	if !ok {
		return nil, fmt.Errorf("event spec must have interface_id")
	}

	interface_id, ok = interface_id.(string)
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
		InterfaceID: interface_id.(string),
		KernelVersion: version.(string),
		Params: params.(*structure.OrderedMap),
	}

	return &spec, nil
}

type Event struct {
	// Publisher 	Identity		`json:"publisher"`
	SpaceID		string			`json:"space_id"`

	Payload		*structure.OrderedMap	`json:"payload"`
	Spec		EventSpec		`json:"spec"`

	Topic		string			`json:"topic"`	
	Subtopic	string			`json:"subtopic"`
	Seperator	string			`json:"seperator"`
	Tag		string			`json:"tag"`
}

func NewEventRequest(spaceID string, payload *structure.OrderedMap, spec *EventSpec, topic string, subtopic string, tag string) *Event{
	return &Event{
			SpaceID: spaceID,
			Payload: payload,
			Spec: *spec,
			Topic: topic,
			Subtopic: subtopic,
			Tag: tag,
	}
}

func (event *Event) Buffer() []byte {
	return []byte(event.String())
}

func (event *Event) String() string {
	return event.Map().Ser()
}

func (event *Event) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()

	// om.Set("publisher", event.Publisher)
	om.Set("space_id", event.SpaceID)
	om.Set("payload", event.Payload)
	om.Set("spec", event.Spec.String())
	om.Set("topic", event.Topic)
	om.Set("subtopic", event.Subtopic)
	om.Set("seperator", event.Seperator)
	om.Set("tag", event.Tag)

	return om
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

	/**
	_, ok = data.Get("publisher")
	if !ok {
		return nil, fmt.Errorf("event data must have key:publisher")
	}
	**/

	specData, ok := data.Get("spec")
	if !ok {
		return nil, fmt.Errorf("event data must have key:spec")
	}

	_, ok = specData.(string)
	if !ok {
		return nil, fmt.Errorf("spec data must be string")
	}

	specData_s, err := structure.ParseOrderedMap(specData.(string))

	if err != nil {
		return nil, fmt.Errorf("spec data type conversion failed %v", err.Error())
	}

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
	Code		int	`json:"code"`
	Message 	string	`json:"message"`
}

func (er *EventExecutionError) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()
	om.Set("code", er.Code)
	om.Set("message", er.Message)
	return om
}

type EventExecutionResult struct {
	Event		*Event			`json:"event"`
	Result		string			`json:"result"`
	Err		*EventExecutionError	`json:"error"`
}

func (xr *EventExecutionResult) Buffer() []byte {
	return []byte(xr.String())
}

func (xr *EventExecutionResult) String() string {
	om := xr.Map()
	return om.Ser()
}

func (xr *EventExecutionResult) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()
	
	// om.Set("event", xr.Event.Map())
	om.Set("result", xr.Result)
	if xr.Err != nil {
		om.Set("error", xr.Err.Map())
	} else {
		om.Set("error", nil)
	}

	return om
}
