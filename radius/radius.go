package radius

import (
	"encoding/binary"
	"errors"
	"net"
	"time"
)

// AttributeType represents the type of an attribute in a RADIUS AVP.
type AttributeType byte

// VendorId represents the vendor ID in a RADIUS AVP.
type VendorId uint32

// avpData represents the data in a RADIUS AVP.
type avpData []byte

// Avp represents a RADIUS Attribute-Value Pair (AVP).
type Avp struct {
	Type     AttributeType
	length   byte
	VendorId VendorId
	Data     avpData
}

// NewAvp creates a new AVP with the given attribute type, vendor ID, and data.
func NewAvp(attributeType AttributeType, vendorId VendorId, avpData avpData) Avp {
	a := Avp{
		Type: attributeType,
		Data: avpData,
	}
	if vendorId == 0 {
		a.length = byte(len(avpData) + 2)
	} else {
		a.VendorId = vendorId
		a.length = byte(len(avpData) + 8)
		a.Data = avpData
	}
	
	return a
}

// NewAvpString creates a new AVP with a string value.
func NewAvpString(attributeType AttributeType, vendorId VendorId, value string) Avp {
	return NewAvp(attributeType, vendorId, []byte(value))
}

// NewAvpUint32 creates a new AVP with a uint32 value.
func NewAvpUint32(attributeType AttributeType, vendorId VendorId, value uint32) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)

	return NewAvp(attributeType, vendorId, buffer)
}

// NewAvpNetIP creates a new AVP with a net.IP value.
func NewAvpNetIP(attributeType AttributeType, vendorId VendorId, value net.IP) Avp {
	return NewAvp(attributeType, vendorId, avpData(value.To4()))
}

// NewAvpTime creates a new AVP with a time.Time value.
func NewAvpTime(attributeType AttributeType, vendorId VendorId, value time.Time) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(value.Unix()))

	return NewAvp(attributeType, vendorId, buffer)
}

// ToBytes converts the AVP to a byte slice.
func (a Avp) ToBytes() []byte {
	bytes := make([]byte, 0)
	if a.VendorId == 0 {
		bytes = append(bytes, byte(a.Type))
		bytes = append(bytes, a.length)
	} else {
		bytes = append(bytes, 26)
		bytes = append(bytes, a.length)
		buffer := make([]byte, 4)
		binary.BigEndian.PutUint32(buffer, uint32(a.VendorId))
		bytes = append(bytes, buffer...)
		bytes = append(bytes, byte(a.Type))
		bytes = append(bytes, byte(len(a.Data)+2))
	}

	bytes = append(bytes, a.Data...)

	return bytes
}

// Avps represents a slice of AVPs.
type Avps []Avp

// NewAvps creates a new slice of AVPs.
func NewAvps() Avps {
	return make(Avps, 0)
}

// Add adds a new AVP to the slice.
func (a Avps) Add(attributeType AttributeType, vendorId VendorId, data avpData) Avps {
	return append(a, NewAvp(attributeType, vendorId, data))
}

// AddAvps adds multiple AVPs to the slice.
func (a Avps) AddAvps(avps ...Avp) Avps {
	return append(a, avps...)
}

// AddString adds a new AVP with a string value to the slice.
func (a Avps) AddString(attributeType AttributeType, vendorId VendorId, value string) Avps {
	return append(a, NewAvpString(attributeType, vendorId, value))
}

// AddUint32 adds a new AVP with a uint32 value to the slice.
func (a Avps) AddUint32(attributeType AttributeType, vendorId VendorId, value uint32) Avps {
	return append(a, NewAvpUint32(attributeType, vendorId, value))
}

// AddNetIP adds a new AVP with a net.IP value to the slice.
func (a Avps) AddNetIP(attributeType AttributeType, vendorId VendorId, value net.IP) Avps {
	return append(a, NewAvpNetIP(attributeType, vendorId, value))
}

// AddTime adds a new AVP with a time.Time value to the slice.
func (a Avps) AddTime(attributeType AttributeType, vendorId VendorId, value time.Time) Avps {
	return append(a, NewAvpTime(attributeType, vendorId, value))
}

// ToBytes converts the slice of AVPs to a byte slice.
func (a Avps) ToBytes() []byte {
	bytes := make([]byte, 0)
	for _, avp := range a {
		bytes = append(bytes, avp.ToBytes()...)
	}

	return bytes
}

// Code represents the code in a RADIUS message.
type Code uint32

// Message represents a RADIUS message.
type Message struct {
	Code          Code
	Identifier    byte
	Authenticator [16]byte
	Avps          Avps
}

// length calculates the length of the RADIUS message.
func (m Message) length() uint16 {
	length := uint16(20)
	for _, avp := range m.Avps {
		length += uint16(avp.length)
	}

	return length
}

// NewMessage creates a new RADIUS message.
func NewMessage(code Code, identifier byte, authenticator [16]byte, avps ...Avp) Message {
	length := uint16(20)
	for _, avp := range avps {
		length += uint16(avp.length)
	}
	
	return Message{
		Code:          code,
		Identifier:    identifier,
		Authenticator: authenticator,
		Avps:          avps,
	}
}

// ToBytes converts the RADIUS message to a byte slice.
func (m Message) ToBytes() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, byte(m.Code))
	bytes = append(bytes, m.Identifier)
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, m.length())
	bytes = append(bytes, buffer...)
	bytes = append(bytes, m.Authenticator[:]...)
	bytes = append(bytes, m.Avps.ToBytes()...)

	return bytes
}

// Get retrieves all AVPs with the given attribute type and vendor ID.
func (a Avps) Get(attributeType AttributeType, vendorId VendorId) []Avp {
	if a == nil {
		return nil
	}

	filteredAvps := NewAvps()

	for _, avp := range a {
		if avp.Type == attributeType && avp.VendorId == vendorId {
			filteredAvps = append(filteredAvps, avp)
		}
	}

	return filteredAvps
}

// GetFirst retrieves the first AVP with the given attribute type and vendor ID.
func (a Avps) GetFirst(attributeType AttributeType, vendorId VendorId) *Avp {
	for _, avp := range a {
		if avp.Type == attributeType && avp.VendorId == vendorId {
			return &avp
		}
	}

	return nil
}

// ToData converts the AVP to a byte slice.
func (a *Avp) ToData() []byte {
	if a == nil {
		return nil
	}

	return a.Data
}

// ToString converts the AVP to a string.
func (a *Avp) ToString() *string {
	if a == nil || a.Data == nil {
		return nil
	}

	value := string(a.Data)

	return &value
}

// ToStringOrDefault converts the AVP to a string or returns a default value.
func (a *Avp) ToStringOrDefault() string {
	value := a.ToString()
	if value == nil {
		var value string
		return value
	}

	return *value
}

// ToUint32 converts the AVP to a uint32.
func (a *Avp) ToUint32() *uint32 {
	if a == nil || a.Data == nil {
		return nil
	}

	value := binary.BigEndian.Uint32(a.Data)

	return &value
}

// ToUint32OrDefault converts the AVP to a uint32 or returns a default value.
func (a *Avp) ToUint32OrDefault() uint32 {
	value := a.ToUint32()
	if value == nil {
		var value uint32
		return value
	}

	return *value
}

// ToNetIP converts the AVP to a net.IP.
func (a *Avp) ToNetIP() *net.IP {
	if a == nil || a.Data == nil {
		return nil
	}

	value := net.IP(a.Data)

	return &value
}

// ToNetIPOrDefault converts the AVP to a net.IP or returns a default value.
func (a *Avp) ToNetIPOrDefault() net.IP {
	value := a.ToNetIP()
	if value == nil {
		var value net.IP
		return value
	}

	return *value
}

// ToTime converts the AVP to a time.Time.
func (a *Avp) ToTime() *time.Time {
	if a == nil || a.Data == nil {
		return nil
	}

	timestamp := int64(binary.BigEndian.Uint32(a.Data))
	value := time.Unix(timestamp, 0)

	return &value
}

// ToTimeOrDefault converts the AVP to a time.Time or returns a default value.
func (a *Avp) ToTimeOrDefault() time.Time {
	value := a.ToTime()
	if value == nil {
		var value time.Time
		return value
	}

	return *value
}

// readAvps reads a byte slice and converts it to a slice of AVPs.
func readAvps(bytes []byte) Avps {
	offset := 0
	avps := NewAvps()

	for offset < len(bytes) {
		attributeType := AttributeType(bytes[offset])
		length := bytes[offset+1]

		var avpData avpData

		var vendorId VendorId
		
		if attributeType == 26 {
			vendorId = VendorId(binary.BigEndian.Uint32(bytes[offset+2 : offset+6]))
			attributeType = AttributeType(bytes[offset+6])
			avpLength := bytes[offset+7]
			avpData = bytes[offset+8 : offset+6+int(avpLength)]
		} else {
			avpLength := bytes[offset+1]
			avpData = bytes[offset+2 : offset+int(avpLength)]
		}
		
		avps = append(avps, NewAvp(attributeType, vendorId, avpData))
		offset += int(length)
	}

	return avps
}

var errInvalidMessageLength = errors.New("invalid message length")

// ReadMessage reads a byte slice and converts it to a RADIUS message.
func ReadMessage(bytes []byte) (*Message, error) {
	if len(bytes) < 20 {
		return nil, errInvalidMessageLength
	}

	authenticator := [16]byte{}
	copy(authenticator[:], bytes[4:20])
	message := Message{
		Code:          Code(bytes[0]),
		Identifier:    bytes[1],
		Authenticator: authenticator,
		Avps:          readAvps(bytes[20:]),
	}

	return &message, nil
}
