package diameter

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
	"time"
)

// Flags represents the flags in a Diameter AVP.
type Flags byte

// Code represents the code in a Diameter AVP.
type Code uint32

// VendorId represents the vendor ID in a Diameter AVP.
type VendorId uint32

// avpData represents the data in a Diameter AVP.
type avpData []byte

// Avp represents a Diameter Attribute-Value Pair (AVP).
type Avp struct {
	Code     Code
	Flags    Flags
	length   uint32
	VendorId VendorId
	Data     avpData
	padding  uint32
}

// WithFlags sets the flags for the AVP.
func (a *Avp) WithFlags(flags Flags) *Avp {
	if a == nil {
		return nil
	}
	a.Flags = flags
	return a
}

// NewAvp creates a new AVP with the given code, flags, vendor ID, and data.
func NewAvp(code Code, flags Flags, vendorId VendorId, avpData avpData) Avp {
	padding := uint32(len(avpData) % 4)
	if padding != 0 {
		padding = 4 - padding
	}
	length := uint32(len(avpData) + 8)
	if vendorId != 0 {
		length += 4
	}
	return Avp{
		Code:     code,
		Flags:    flags,
		length:   length,
		VendorId: vendorId,
		Data:     avpData,
		padding:  padding,
	}
}

// NewAvpGroup creates a new grouped AVP with the given code, flags, vendor ID, and AVPs.
func NewAvpGroup(code Code, flags Flags, vendorId VendorId, avps ...Avp) Avp {
	_avps := Avps(avps)
	return NewAvp(code, flags, vendorId, _avps.ToBytes())
}

// NewAvpString creates a new AVP with a string value.
func NewAvpString(code Code, flags Flags, vendorId VendorId, value string) Avp {
	return NewAvp(code, flags, vendorId, []byte(value))
}

// NewAvpUint32 creates a new AVP with a uint32 value.
func NewAvpUint32(code Code, flags Flags, vendorId VendorId, value uint32) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)
	return NewAvp(code, flags, vendorId, buffer)
}

// NewAvpUint64 creates a new AVP with a uint64 value.
func NewAvpUint64(code Code, flags Flags, vendorId VendorId, value uint64) Avp {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, value)
	return NewAvp(code, flags, vendorId, buffer)
}

// NewAvpFloat32 creates a new AVP with a float32 value.
func NewAvpFloat32(code Code, flags Flags, vendorId VendorId, value float32) Avp {
	bits := math.Float32bits(value)
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, bits)
	return NewAvp(code, flags, vendorId, buffer)
}

// NewAvpFloat64 creates a new AVP with a float64 value.
func NewAvpFloat64(code Code, flags Flags, vendorId VendorId, value float64) Avp {
	bits := math.Float64bits(value)
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, bits)
	return NewAvp(code, flags, vendorId, buffer)
}

// NewAvpNetIP creates a new AVP with a net.IP value.
func NewAvpNetIP(code Code, flags Flags, vendorId VendorId, value net.IP) Avp {
	if value.To4() != nil {
		avpData := make([]byte, 6)
		avpData[1] = 1
		copy(avpData[2:], value.To4())
		return NewAvp(code, flags, vendorId, avpData)
	} else {
		avpData := make([]byte, 18)
		avpData[1] = 2
		copy(avpData[2:], value.To16())
		return NewAvp(code, flags, vendorId, avpData)
	}
}

// NewAvpTime creates a new AVP with a time.Time value.
func NewAvpTime(code Code, flags Flags, vendorId VendorId, value time.Time) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(value.Unix()))
	return NewAvp(code, flags, vendorId, buffer)
}

// ToBytes converts the AVP to a byte slice.
func (a Avp) ToBytes() []byte {
	bytes := make([]byte, a.length+a.padding)
	binary.BigEndian.PutUint32(bytes, uint32(a.Code))
	bytes[4] = byte(a.Flags)
	copy(bytes[5:8], writeUInt24(a.length))
	if a.VendorId != 0 {
		binary.BigEndian.PutUint32(bytes[8:12], uint32(a.VendorId))
		copy(bytes[12:], a.Data)
	} else {
		copy(bytes[8:], a.Data)
	}
	return bytes
}

// Avps represents a slice of AVPs.
type Avps []Avp

// NewAvps creates a new slice of AVPs.
func NewAvps() Avps {
	return make([]Avp, 0)
}

// ToBytes converts the slice of AVPs to a byte slice.
func (a Avps) ToBytes() []byte {
	bytes := make([]byte, 0)
	for _, avp := range a {
		bytes = append(bytes, avp.ToBytes()...)
	}
	return bytes
}

// Add adds a new AVP to the slice.
func (a Avps) Add(code Code, flags Flags, vendorId VendorId, data avpData) Avps {
	return append(a, NewAvp(code, flags, vendorId, data))
}

// AddAvps adds multiple AVPs to the slice.
func (a Avps) AddAvps(avps ...Avp) Avps {
	return append(a, avps...)
}

// AddString adds a new AVP with a string value to the slice.
func (a Avps) AddString(code Code, flags Flags, vendorId VendorId, value string) Avps {
	return append(a, NewAvpString(code, flags, vendorId, value))
}

// AddUint32 adds a new AVP with a uint32 value to the slice.
func (a Avps) AddUint32(code Code, flags Flags, vendorId VendorId, value uint32) Avps {
	return append(a, NewAvpUint32(code, flags, vendorId, value))
}

// AddUint64 adds a new AVP with a uint64 value to the slice.
func (a Avps) AddUint64(code Code, flags Flags, vendorId VendorId, value uint64) Avps {
	return append(a, NewAvpUint64(code, flags, vendorId, value))
}

// AddFloat32 adds a new AVP with a float32 value to the slice.
func (a Avps) AddFloat32(code Code, flags Flags, vendorId VendorId, value float32) Avps {
	return append(a, NewAvpFloat32(code, flags, vendorId, value))
}

// AddFloat64 adds a new AVP with a float64 value to the slice.
func (a Avps) AddFloat64(code Code, flags Flags, vendorId VendorId, value float64) Avps {
	return append(a, NewAvpFloat64(code, flags, vendorId, value))
}

// AddNetIP adds a new AVP with a net.IP value to the slice.
func (a Avps) AddNetIP(code Code, flags Flags, vendorId VendorId, value net.IP) Avps {
	return append(a, NewAvpNetIP(code, flags, vendorId, value))
}

// AddTime adds a new AVP with a time.Time value to the slice.
func (a Avps) AddTime(code Code, flags Flags, vendorId VendorId, value time.Time) Avps {
	return append(a, NewAvpTime(code, flags, vendorId, value))
}

// AddGroup adds a new grouped AVP to the slice.
func (a Avps) AddGroup(code Code, flags Flags, vendorId VendorId, groupAvps ...Avp) Avps {
	return append(a, NewAvpGroup(code, flags, vendorId, groupAvps...))
}

// ApplicationId represents the application ID in a Diameter message.
type ApplicationId uint32

// toBytes converts the ApplicationId to a byte slice.
func (a ApplicationId) toBytes() []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(a))
	return bytes
}

// writeUInt24 writes a uint32 value as a 3-byte slice.
func writeUInt24(value uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, value)
	return bytes[1:]
}

// CommandCode represents the command code in a Diameter message.
type CommandCode uint32

// toBytes converts the CommandCode to a byte slice.
func (c CommandCode) toBytes() []byte {
	return writeUInt24(uint32(c))
}

// Message represents a Diameter message.
type Message struct {
	Version       byte
	Flags         Flags
	CommandCode   CommandCode
	ApplicationId ApplicationId
	HopByHopId    [4]byte
	EndToEndId    [4]byte
	Avps          Avps
}

// length calculates the length of the Diameter message.
func (m Message) length() uint32 {
	length := uint32(20)
	for _, avp := range m.Avps {
		length += avp.length + avp.padding
	}
	return length
}

// NewMessage creates a new Diameter message.
func NewMessage(version byte, flags Flags, commandCode CommandCode, applicationId ApplicationId, hopByHopId [4]byte, endToEndId [4]byte, avps ...Avp) Message {
	return Message{
		Version:       version,
		Flags:         flags,
		CommandCode:   commandCode,
		ApplicationId: applicationId,
		HopByHopId:    hopByHopId,
		EndToEndId:    endToEndId,
		Avps:          avps,
	}
}

// ToBytes converts the Diameter message to a byte slice.
func (message Message) ToBytes() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, message.Version)
	bytes = append(bytes, writeUInt24(message.length())...)
	bytes = append(bytes, byte(message.Flags))
	bytes = append(bytes, message.CommandCode.toBytes()...)
	bytes = append(bytes, message.ApplicationId.toBytes()...)
	bytes = append(bytes, message.HopByHopId[:]...)
	bytes = append(bytes, message.EndToEndId[:]...)
	bytes = append(bytes, message.Avps.ToBytes()...)
	return bytes
}

// Get retrieves all AVPs with the given code and vendor ID.
func (a Avps) Get(code Code, vendorId VendorId) Avps {
	filteredAvps := NewAvps()
	for _, avp := range a {
		if avp.Code == code && avp.VendorId == vendorId {
			filteredAvps = append(filteredAvps, avp)
		}
	}
	return filteredAvps
}

// GetFirst retrieves the first AVP with the given code and vendor ID.
func (a Avps) GetFirst(code Code, vendorId VendorId) *Avp {
	for _, avp := range a {
		if avp.Code == code && avp.VendorId == vendorId {
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

// ToUint64 converts the AVP to a uint64.
func (a *Avp) ToUint64() *uint64 {
	if a == nil || a.Data == nil {
		return nil
	}
	value := binary.BigEndian.Uint64(a.Data)
	return &value
}

// ToUint64OrDefault converts the AVP to a uint64 or returns a default value.
func (a *Avp) ToUint64OrDefault() uint64 {
	value := a.ToUint64()
	if value == nil {
		var value uint64
		return value
	}
	return *value
}

// ToFloat32 converts the AVP to a float32.
func (a *Avp) ToFloat32() *float32 {
	if a == nil || a.Data == nil {
		return nil
	}
	bits := binary.BigEndian.Uint32(a.Data)
	value := math.Float32frombits(bits)
	return &value
}

// ToFloat32OrDefault converts the AVP to a float32 or returns a default value.
func (a *Avp) ToFloat32OrDefault() float32 {
	value := a.ToFloat32()
	if value == nil {
		var value float32
		return value
	}
	return *value
}

// ToFloat64 converts the AVP to a float64.
func (a *Avp) ToFloat64() *float64 {
	if a == nil || a.Data == nil {
		return nil
	}
	bits := binary.BigEndian.Uint64(a.Data)
	value := math.Float64frombits(bits)
	return &value
}

// ToFloat64OrDefault converts the AVP to a float64 or returns a default value.
func (a *Avp) ToFloat64OrDefault() float64 {
	value := a.ToFloat64()
	if value == nil {
		var value float64
		return value
	}
	return *value
}

// ToNetIP converts the AVP to a net.IP.
func (a *Avp) ToNetIP() *net.IP {
	if a == nil || a.Data == nil {
		return nil
	}
	if a.Data[1] == 1 {
		value := net.IP(a.Data[2:6])
		return &value
	} else {
		value := net.IP(a.Data[2:18])
		return &value
	}
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
	value := time.Unix(timestamp-2208988800, 0)
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

// ToGroup converts the AVP to a grouped AVP.
func (a *Avp) ToGroup() Avps {
	if a == nil || a.Data == nil {
		return NewAvps()
	}
	return readAvps(a.Data)
}

// readAvps reads a byte slice and converts it to a slice of AVPs.
func readAvps(bytes []byte) Avps {
	offset := 0
	avps := NewAvps()
	for offset < len(bytes) {
		code := Code(binary.BigEndian.Uint32(bytes[offset : offset+4]))
		flags := Flags(bytes[offset+4])
		vendorSpecific := flags&0x80 != 0
		length := int(readUInt24(bytes[offset+5 : offset+8]))
		var vendorId VendorId
		var avpData avpData
		if vendorSpecific {
			vendorId = VendorId(binary.BigEndian.Uint32(bytes[offset+8 : offset+12]))
			avpData = bytes[offset+12 : offset+length]
		} else {
			avpData = bytes[offset+8 : offset+length]
		}
		avp := NewAvp(code, flags, vendorId, avpData)
		avps = append(avps, avp)
		offset += length + int(avp.padding)
	}
	return avps
}

// readUInt24 reads a 3-byte slice and converts it to a uint32.
func readUInt24(bytes []byte) uint32 {
	if len(bytes) == 3 {
		bytes = append([]byte{0}, bytes[:]...)
	}
	return binary.BigEndian.Uint32(bytes)
}

// ReadMessage reads a byte slice and converts it to a Diameter message.
func ReadMessage(bytes []byte) (*Message, error) {
	if len(bytes) < 20 {
		return nil, errors.New("invalid message length")
	}
	hopByHopId := [4]byte{}
	copy(hopByHopId[:], bytes[12:16])
	endToEndId := [4]byte{}
	copy(endToEndId[:], bytes[16:20])
	message := Message{
		Version:       bytes[0],
		Flags:         Flags(bytes[4]),
		CommandCode:   CommandCode(readUInt24(bytes[5:8])),
		ApplicationId: ApplicationId(binary.BigEndian.Uint32(bytes[8:12])),
		HopByHopId:    hopByHopId,
		EndToEndId:    endToEndId,
		Avps:          readAvps(bytes[20:]),
	}
	return &message, nil
}
