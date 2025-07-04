package model

import (
	"fortuna/structure"
)

type Identity struct {
	OrganizationID		string  `json:"organization_id"`
	WorkspaceID		string	`json:"workspace_id"`
	UserID			string	`json:"user_id"`
}

type EventInterface struct {
	InterfaceID	int64		`json:"protocol_id"`
	InterfaceName	string		`json:"protocol_name"`
}

type EventSpec struct {
	Protocol	EventInterface		`json:"protocol"`
	// KernelVersion	KernelVersion		`json:"kernel_version"`
	Params		*structure.OrderedMap	`json:"param0"`
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

type EventExecutionError struct {
	Code		int
	Message 	string
}

type EventExecutionResult struct {
	Event		*Event
	Result		string
	Err		*EventExecutionError
}


