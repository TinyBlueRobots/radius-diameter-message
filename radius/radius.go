package radius

import (
	"encoding/binary"
	"net"
	"time"
)

type AttributeType byte
type VendorId uint32
type avpData []byte

type Avp struct {
	Type     AttributeType
	length   byte
	VendorId VendorId
	Data     avpData
}

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

func NewAvpString(attributeType AttributeType, vendorId VendorId, value string) Avp {
	return NewAvp(attributeType, vendorId, []byte(value))
}

func NewAvpUint32(attributeType AttributeType, vendorId VendorId, value uint32) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)
	return NewAvp(attributeType, vendorId, buffer)
}

func NewAvpNetIP(attributeType AttributeType, vendorId VendorId, value net.IP) Avp {
	return NewAvp(attributeType, vendorId, avpData(value.To4()))
}

func NewAvpTime(attributeType AttributeType, vendorId VendorId, value time.Time) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(value.Unix()))
	return NewAvp(attributeType, vendorId, buffer)
}

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

type Avps []Avp

func NewAvps() Avps {
	return make(Avps, 0)
}

func (a Avps) Add(attributeType AttributeType, vendorId VendorId, data avpData) Avps {
	return append(a, NewAvp(attributeType, vendorId, data))
}

func (a Avps) AddAvp(avps ...Avp) Avps {
	return append(a, avps...)
}

func (a Avps) AddString(attributeType AttributeType, vendorId VendorId, value string) Avps {
	return append(a, NewAvpString(attributeType, vendorId, value))
}

func (a Avps) AddUint32(attributeType AttributeType, vendorId VendorId, value uint32) Avps {
	return append(a, NewAvpUint32(attributeType, vendorId, value))
}

func (a Avps) AddNetIP(attributeType AttributeType, vendorId VendorId, value net.IP) Avps {
	return append(a, NewAvpNetIP(attributeType, vendorId, value))
}

func (a Avps) AddTime(attributeType AttributeType, vendorId VendorId, value time.Time) Avps {
	return append(a, NewAvpTime(attributeType, vendorId, value))
}

func (a Avps) ToBytes() []byte {
	bytes := make([]byte, 0)
	for _, avp := range a {
		bytes = append(bytes, avp.ToBytes()...)
	}
	return bytes
}

type Code uint32

type Message struct {
	Code          Code
	Identifier    byte
	Authenticator [16]byte
	Avps          Avps
}

func (m Message) length() uint16 {
	length := uint16(20)
	for _, avp := range m.Avps {
		length += uint16(avp.length)
	}
	return length
}

func NewMessage(code Code, identifier byte, authenticator [16]byte, avps Avps) Message {
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

func (a Avps) GetFirst(attributeType AttributeType, vendorId VendorId) *Avp {
	for _, avp := range a {
		if avp.Type == attributeType && avp.VendorId == vendorId {
			return &avp
		}
	}
	return nil
}

func (a *Avp) ToString() *string {
	if a == nil || a.Data == nil {
		return nil
	}
	value := string(a.Data)
	return &value
}

func (a *Avp) ToStringOrDefault() string {
	value := a.ToString()
	if value == nil {
		var value string
		return value
	}
	return *value
}

func (a *Avp) ToUint32() *uint32 {
	if a == nil || a.Data == nil {
		return nil
	}
	value := binary.BigEndian.Uint32(a.Data)
	return &value
}

func (a *Avp) ToUint32OrDefault() uint32 {
	value := a.ToUint32()
	if value == nil {
		var value uint32
		return value
	}
	return *value
}

func (a *Avp) ToNetIP() *net.IP {
	if a == nil || a.Data == nil {
		return nil
	}
	value := net.IP(a.Data)
	return &value
}

func (a *Avp) ToNetIPOrDefault() net.IP {
	value := a.ToNetIP()
	if value == nil {
		var value net.IP
		return value
	}
	return *value
}

func (a *Avp) ToTime() *time.Time {
	if a == nil || a.Data == nil {
		return nil
	}
	timestamp := int64(binary.BigEndian.Uint32(a.Data))
	value := time.Unix(timestamp, 0)
	return &value
}

func (a *Avp) ToTimeOrDefault() time.Time {
	value := a.ToTime()
	if value == nil {
		var value time.Time
		return value
	}
	return *value
}

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

func ReadMessage(bytes []byte) *Message {
	if len(bytes) < 20 {
		return nil
	}
	authenticator := [16]byte{}
	copy(authenticator[:], bytes[4:20])
	message := Message{
		Code:          Code(bytes[0]),
		Identifier:    bytes[1],
		Authenticator: authenticator,
		Avps:          readAvps(bytes[20:]),
	}
	return &message
}
