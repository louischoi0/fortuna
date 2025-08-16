package model

import (
	"bytes"
	"fmt"
	C "fortuna/core/config"
	"fortuna/crypto"
	"fortuna/structure"
	"fortuna/util"
	"strconv"
	"strings"
)

type Identity struct {
	Address        string `json:"address"`
	OrganizationID string `json:"organization_id"`
	WorkspaceID    string `json:"workspace_id"`
	UserID         string `json:"user_id"`
}

type EventSpec struct {
	InterfaceID   string                `json:"interface_id"`
	KernelVersion string                `json:"kernel_version"`
	Params        *structure.OrderedMap `json:"params"`
}

func NewEventSpec(interface_id string, kernel_version string, params *structure.OrderedMap) *EventSpec {
	return &EventSpec{
		InterfaceID:   interface_id,
		KernelVersion: kernel_version,
		Params:        params,
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

func NewEventSpecFromOrderedMap(data *structure.OrderedMap) (*EventSpec, error) {
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
		InterfaceID:   interface_id.(string),
		KernelVersion: version.(string),
		Params:        params.(*structure.OrderedMap),
	}

	return &spec, nil
}

type Event struct {
	Timestamp uint64 `json:"timestamp"`
	Publisher string `json:"publisher"`
	SpaceID   string `json:"space_id"`

	Payload *structure.OrderedMap `json:"payload"`
	Spec    *EventSpec            `json:"spec"`

	Topic     string `json:"topic"`
	Subtopic  string `json:"subtopic"`
	Seperator string `json:"seperator"`
	Tag       string `json:"tag"`
}

// DecodeEvent decodes bytes produced by (*Event).Encode() back into an Event.
// It also verifies the leading hash matches event.Hash() after reconstruction.
func DecodeEvent(b []byte) (*Event, error) {
	var (
		off = 0
		n   = len(b)
	)

	need := func(k int) error {
		if off+k > n {
			return fmt.Errorf("buffer underflow: %d", k)
		}
		return nil
	}
	readFixedString := func(k int) (string, error) {
		if err := need(k); err != nil {
			return "", err
		}
		s := string(b[off : off+k])
		off += k
		return s, nil
	}
	readU64LE := func() (uint64, error) {
		if err := need(8); err != nil {
			return 0, err
		}
		u, err := util.DecodeUint64(b[off : off+8])
		if err != nil {
			return 0, err
		}
		off += 8
		return u, nil
	}

	hash, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read hash: %w", err)
	}

	timestamp, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read timestamp: %w", err)
	}

	publisher, err := readFixedString(C.IDENTITY_ADDRESS_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read publisher: %w", err)
	}

	spaceID, err := readFixedString(C.SPACE_ID_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read spaceID: %w", err)
	}
	ifaceID, err := readFixedString(C.INTERFACE_ID_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read interfaceID: %w", err)
	}
	kernelVer, err := readFixedString(C.KERNEL_VERSION_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read kernelVersion: %w", err)
	}

	if err := need(1); err != nil {
		return nil, fmt.Errorf("read params flag: %w", err)
	}
	paramsStart := off
	off++

	totalLenParams, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read params totalLen: %w", err)
	}
	if err := need(int(totalLenParams)); err != nil {
		return nil, fmt.Errorf("params body underflow: %w", err)
	}
	paramsBlob := b[paramsStart : paramsStart+1+8+int(totalLenParams)]
	off += int(totalLenParams)

	params, err := util.DecodeOrderedMap(paramsBlob)
	if err != nil {
		return nil, fmt.Errorf("DecodeOrderedMap(params): %w", err)
	}

	if err := need(1); err != nil {
		return nil, fmt.Errorf("read payload flag: %w", err)
	}
	payloadStart := off
	off++

	totalLenPayload, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read payload totalLen: %w", err)
	}
	if err := need(int(totalLenPayload)); err != nil {
		return nil, fmt.Errorf("payload body underflow: %w", err)
	}
	payloadBlob := b[payloadStart : payloadStart+1+8+int(totalLenPayload)]
	off += int(totalLenPayload)

	payload, err := util.DecodeOrderedMap(payloadBlob)
	if err != nil {
		return nil, fmt.Errorf("DecodeOrderedMap(payload): %w", err)
	}

	topicRaw, err := readFixedString(16)
	if err != nil {
		return nil, fmt.Errorf("read topic: %w", err)
	}
	subtopicRaw, err := readFixedString(16)
	if err != nil {
		return nil, fmt.Errorf("read subtopic: %w", err)
	}
	sepRaw, err := readFixedString(16)
	if err != nil {
		return nil, fmt.Errorf("read seperator: %w", err)
	}
	tagRaw, err := readFixedString(16)
	if err != nil {
		return nil, fmt.Errorf("read tag: %w", err)
	}

	evt := &Event{
		Timestamp: timestamp,
		Publisher: publisher,
		SpaceID:   spaceID,
		Spec: &EventSpec{
			InterfaceID:   ifaceID,
			KernelVersion: kernelVer,
			Params:        params,
		},
		Payload:   payload,
		Topic:     util.UnpadLeftS(topicRaw),
		Subtopic:  util.UnpadLeftS(subtopicRaw),
		Seperator: util.UnpadLeftS(sepRaw),
		Tag:       util.UnpadLeftS(tagRaw),
	}

	if got := evt.Hash(); got != hash {
		return nil, fmt.Errorf("hash mismatch: header=%q computed=%q", hash, got)
	}

	return evt, nil
}

func (event *Event) Encode() ([]byte, error) {
	var buf bytes.Buffer

	hash := event.Hash()
	if len(hash) != C.MODEL_HASH_STR_LENGTH {
		return nil, fmt.Errorf("hash must have length: %d", C.MODEL_HASH_STR_LENGTH)
	}

	if len(event.SpaceID) != C.SPACE_ID_STR_LENGTH {
		return nil, fmt.Errorf("spaceID must have length: %d", C.SPACE_ID_STR_LENGTH)
	}

	if len(event.Publisher) != C.IDENTITY_ADDRESS_STR_LENGTH {
		return nil, fmt.Errorf("publisher must have length: %d", C.IDENTITY_ADDRESS_STR_LENGTH)
	}

	if len(event.Spec.InterfaceID) != C.INTERFACE_ID_STR_LENGTH {
		return nil, fmt.Errorf("interfaceID must have length: %d", C.INTERFACE_ID_STR_LENGTH)
	}

	if len(event.Spec.KernelVersion) != C.KERNEL_VERSION_STR_LENGTH {
		return nil, fmt.Errorf("kernelVersion must have length: %d", C.KERNEL_VERSION_STR_LENGTH)
	}

	buf.WriteString(hash)
	buf.Write(util.EncodeUint64(event.Timestamp))
	buf.WriteString(event.Publisher)
	buf.WriteString(event.SpaceID)
	buf.WriteString(event.Spec.InterfaceID)
	buf.WriteString(event.Spec.KernelVersion)

	parambuf, err := util.EncodeOrderedMap(event.Spec.Params)
	if err != nil {
		return nil, err
	}

	buf.Write(parambuf)

	payloadbuf, err := util.EncodeOrderedMap(event.Payload)
	if err != nil {
		return nil, err
	}

	buf.Write(payloadbuf)
	buf.WriteString(util.PadLeftS(event.Topic, 16))
	buf.WriteString(util.PadLeftS(event.Subtopic, 16))
	buf.WriteString(util.PadLeftS(event.Seperator, 16))
	buf.WriteString(util.PadLeftS(event.Tag, 16))

	return buf.Bytes(), nil
}

func (event *Event) Verify(eventHash string) bool {
	return event.Hash() == eventHash
}

func (event *Event) Hash() string {
	var buf strings.Builder

	buf.WriteString(strconv.Itoa(int(event.Timestamp)))
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(event.SpaceID)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(event.Spec.InterfaceID)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(event.Spec.KernelVersion)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(event.Publisher)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(event.Spec.Params.Hash())
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(event.Payload.Hash())

	return crypto.SHA256(buf.String())
}

func NewEventRequest(publisher string, spaceID string, payload *structure.OrderedMap, spec *EventSpec, topic string, subtopic string, tag string) (*Event, error) {

	if len(publisher) != C.IDENTITY_ADDRESS_STR_LENGTH {
		return nil, fmt.Errorf("publisher must have length: %d", C.IDENTITY_ADDRESS_STR_LENGTH)
	}

	if len(spaceID) != C.SPACE_ID_STR_LENGTH {
		return nil, fmt.Errorf("spaceID must have length: %d", C.SPACE_ID_STR_LENGTH)
	}

	return &Event{
		Publisher: publisher,
		SpaceID:   spaceID,
		Payload:   payload,
		Spec:      spec,
		Topic:     topic,
		Subtopic:  subtopic,
		Tag:       tag,
	}, nil
}

func (event *Event) Buffer() []byte {
	return []byte(event.String())
}

func (event *Event) String() string {
	return event.Map().Ser()
}

func (event *Event) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()

	om.Set("publisher", event.Publisher)
	om.Set("space_id", event.SpaceID)
	om.Set("payload", event.Payload)
	om.Set("spec", event.Spec.Map())
	om.Set("topic", event.Topic)
	om.Set("subtopic", event.Subtopic)
	om.Set("seperator", event.Seperator)
	om.Set("tag", event.Tag)

	return om
}

func NewEventFromOrderedMap(data *structure.OrderedMap) (*Event, error) {

	publisher, ok := data.String("publisher")
	if !ok {
		return nil, fmt.Errorf("event data must have key:publisher")
	}

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

	if len(spaceID.(string)) != C.SPACE_ID_STR_LENGTH {
		return nil, fmt.Errorf("spaceID must have length: %d", C.SPACE_ID_STR_LENGTH)
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

	_, ok = specData.(*structure.OrderedMap)

	if !ok {
		return nil, fmt.Errorf("event spec data must have ordered map")
	}

	spec, err := NewEventSpecFromOrderedMap(specData.(*structure.OrderedMap))
	if err != nil {
		return nil, fmt.Errorf("event spec data must have ordered map")
	}

	topic, ok := data.Get("topic")
	if !ok {
		return nil, fmt.Errorf("event data must have key:topic")
	}

	topic_s, ok := topic.(string)
	if !ok {
		return nil, fmt.Errorf("topic must have type: string")
	}

	event := Event{
		Timestamp: util.GetCurrentTime(),
		Publisher: publisher,
		SpaceID:   spaceID_s,
		Spec:      spec,
		Payload:   payload_s,
		Topic:     topic_s,
	}

	return &event, nil
}

type EventExecutionError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (er *EventExecutionError) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()
	om.Set("code", er.Code)
	om.Set("message", er.Message)
	return om
}

type EventResult struct {
	Event     *Event               `json:"event"`
	EventHash string               `json:"event_hash"`
	Result    string               `json:"result"`
	Err       *EventExecutionError `json:"error"`
}

func (xr *EventResult) GetSpaceID() string {
	return xr.Event.SpaceID
}

func (xr *EventResult) Buffer() []byte {
	return []byte(xr.String())
}

func (xr *EventResult) String() string {
	om := xr.Map()
	return om.Ser()
}

func (xr *EventResult) Hash() string {
	if xr.Event == nil {
		return ""
	}

	var buffer strings.Builder
	buffer.WriteString(xr.Event.Hash())
	buffer.WriteString(C.HASH_SEPERATOR)
	buffer.WriteString(xr.Result)

	return crypto.SHA256(buffer.String())
}

func NewEventResultFromEvent(event *Event, result string) *EventResult {
	return &EventResult{
		Event:     event,
		Result:    result,
		EventHash: event.Hash(),
		Err:       nil,
	}
}

func (xr *EventResult) Encode() ([]byte, error) {
	if xr.Event == nil {
		return nil, fmt.Errorf("execution has no event")
	}

	var buf bytes.Buffer

	buf.WriteString(xr.Hash())

	eventBuffer, err := xr.Event.Encode()
	if err != nil {
		return nil, err
	}

	eventBufferSize := uint64(len(eventBuffer))
	buf.Write(util.EncodeUint64(eventBufferSize))
	buf.Write(eventBuffer)

	resultSize := uint64(len(xr.Result))

	buf.Write(util.EncodeUint64(resultSize))
	buf.Write([]byte(xr.Result))

	return buf.Bytes(), nil
}

func (xr *EventResult) Verify(hash string) bool {
	return xr.Hash() == hash
}

func (xr *EventResult) Map() *structure.OrderedMap {
	om := structure.NewOrderedMap()

	if xr.Event != nil {
		om.Set("event", xr.Event.Map())
		om.Set("event_hash", xr.Event.Hash())
	} else {
		om.Set("event", nil)
		om.Set("event_hash", nil)
	}

	om.Set("result", xr.Result)

	if xr.Err != nil {
		om.Set("error", xr.Err.Map())
	} else {
		om.Set("error", nil)
	}

	return om
}

func DecodeEventResult(b []byte) (*EventResult, error) {
	var (
		off = 0
		n   = len(b)
	)

	need := func(k int) error {
		if off+k > n {
			return fmt.Errorf("buffer underflow: need %d bytes", k)
		}
		return nil
	}
	readFixedString := func(k int) (string, error) {
		if err := need(k); err != nil {
			return "", err
		}
		s := string(b[off : off+k])
		off += k
		return s, nil
	}
	readU64LE := func() (uint64, error) {
		if err := need(8); err != nil {
			return 0, err
		}
		v, err := util.DecodeUint64(b[off : off+8])
		if err != nil {
			return 0, err
		}
		off += 8
		return v, nil
	}

	// 1. Read hash
	hash, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read hash: %w", err)
	}

	// 2. Read event buffer size
	eventSize, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read event size: %w", err)
	}

	if err := need(int(eventSize)); err != nil {
		return nil, fmt.Errorf("read event buffer: %w", err)
	}
	eventBuf := b[off : off+int(eventSize)]
	off += int(eventSize)

	event, err := DecodeEvent(eventBuf)
	if err != nil {
		return nil, fmt.Errorf("decode event: %w", err)
	}

	// 3. Read result size
	resultSize, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read result size: %w", err)
	}

	if err := need(int(resultSize)); err != nil {
		return nil, fmt.Errorf("read result buffer: %w", err)
	}
	resultBytes := b[off : off+int(resultSize)]
	result := string(resultBytes)

	event_hash := event.Hash()

	// 4. Compose object
	exec := &EventResult{
		Event:     event,
		EventHash: event_hash,
		Result:    result,
		Err:       nil, // not included in current encoding
	}

	if hash != exec.Hash() {
		return nil, fmt.Errorf("event result hash mismatched expected: %s, got: %s", hash, exec.Hash())
	}

	return exec, nil
}
