package diameter

import (
	"encoding/binary"
	"math"
	"net"
	"time"
)

type Flags byte
type Code uint32
type VendorId uint32
type avpData []byte

type Avp struct {
	Code     Code
	Flags    Flags
	length   uint32
	VendorId VendorId
	Data     avpData
	padding  uint32
}

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

func NewAvpGroup(code Code, flags Flags, vendorId VendorId, avps Avps) Avp {
	return NewAvp(code, flags, vendorId, avps.ToBytes())
}

func NewAvpString(code Code, flags Flags, vendorId VendorId, value string) Avp {
	return NewAvp(code, flags, vendorId, []byte(value))
}

func NewAvpUint32(code Code, flags Flags, vendorId VendorId, value uint32) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)
	return NewAvp(code, flags, vendorId, buffer)
}

func NewAvpUint64(code Code, flags Flags, vendorId VendorId, value uint64) Avp {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, value)
	return NewAvp(code, flags, vendorId, buffer)
}

func NewAvpFloat32(code Code, flags Flags, vendorId VendorId, value float32) Avp {
	bits := math.Float32bits(value)
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, bits)
	return NewAvp(code, flags, vendorId, buffer)
}

func NewAvpFloat64(code Code, flags Flags, vendorId VendorId, value float64) Avp {
	bits := math.Float64bits(value)
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, bits)
	return NewAvp(code, flags, vendorId, buffer)
}

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

func NewAvpTime(code Code, flags Flags, vendorId VendorId, value time.Time) Avp {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(value.Unix()))
	return NewAvp(code, flags, vendorId, buffer)
}

func (avp Avp) ToBytes() []byte {
	bytes := make([]byte, avp.length+avp.padding)
	binary.BigEndian.PutUint32(bytes, uint32(avp.Code))
	bytes[4] = byte(avp.Flags)
	copy(bytes[5:8], writeUInt24(avp.length))
	copy(bytes[8:], avp.Data)
	return bytes
}

type Avps []Avp

func NewAvps() Avps {
	return make([]Avp, 0)
}

func (avps Avps) ToBytes() []byte {
	bytes := make([]byte, 0)
	for _, avp := range avps {
		bytes = append(bytes, avp.ToBytes()...)
	}
	return bytes
}

func (avps Avps) Add(code Code, vendorId VendorId, flags Flags, data avpData) Avps {
	return append(avps, NewAvp(code, flags, vendorId, data))
}

func (avps Avps) AddAvp(avp Avp) Avps {
	return append(avps, avp)
}

func (avps Avps) AddString(code Code, vendorId VendorId, flags Flags, value string) Avps {
	return append(avps, NewAvpString(code, flags, vendorId, value))
}

func (avps Avps) AddUint32(code Code, vendorId VendorId, flags Flags, value uint32) Avps {
	return append(avps, NewAvpUint32(code, flags, vendorId, value))
}

func (avps Avps) AddUint64(code Code, vendorId VendorId, flags Flags, value uint64) Avps {
	return append(avps, NewAvpUint64(code, flags, vendorId, value))
}

func (avps Avps) AddFloat32(code Code, vendorId VendorId, flags Flags, value float32) Avps {
	return append(avps, NewAvpFloat32(code, flags, vendorId, value))
}

func (avps Avps) AddFloat64(code Code, vendorId VendorId, flags Flags, value float64) Avps {
	return append(avps, NewAvpFloat64(code, flags, vendorId, value))
}

func (avps Avps) AddNetIP(code Code, vendorId VendorId, flags Flags, value net.IP) Avps {
	return append(avps, NewAvpNetIP(code, flags, vendorId, value))
}

func (avps Avps) AddTime(code Code, vendorId VendorId, flags Flags, value time.Time) Avps {
	return append(avps, NewAvpTime(code, flags, vendorId, value))
}

func (avps Avps) AddGroup(code Code, vendorId VendorId, flags Flags, groupAvps Avps) Avps {
	return append(avps, NewAvpGroup(code, flags, vendorId, groupAvps))
}

type ApplicationId uint32

func (applicationId ApplicationId) toBytes() []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(applicationId))
	return bytes
}

func writeUInt24(value uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, value)
	return bytes[1:]
}

type CommandCode uint32

func (commandCode CommandCode) toBytes() []byte {
	return writeUInt24(uint32(commandCode))
}

type Message struct {
	Version       byte
	Flags         Flags
	CommandCode   CommandCode
	ApplicationId ApplicationId
	HopByHopId    [4]byte
	EndToEndId    [4]byte
	Avps          Avps
}

func (message Message) length() uint32 {
	length := uint32(20)
	for _, avp := range message.Avps {
		length += avp.length + avp.padding
	}
	return length
}

func NewMessage(version byte, flags Flags, commandCode CommandCode, applicationId ApplicationId, hopByHopId [4]byte, endToEndId [4]byte, avps Avps) Message {
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

func (avps Avps) Get(code Code, vendorId VendorId) Avps {
	if avps == nil {
		return nil
	}
	filteredAvps := NewAvps()
	for _, avp := range avps {
		if avp.Code == code && avp.VendorId == vendorId {
			filteredAvps = append(filteredAvps, avp)
		}
	}
	return filteredAvps
}

func (avps Avps) GetFirst(code Code, vendorId VendorId) *Avp {
	if avps == nil {
		return nil
	}
	for _, avp := range avps {
		if avp.Code == code && avp.VendorId == vendorId {
			return &avp
		}
	}
	return nil
}

func (avp Avp) ToString() *string {
	if avp.Data == nil {
		return nil
	}

	value := string(avp.Data)
	return &value
}

func (avp Avp) ToUint32() *uint32 {
	if avp.Data == nil {
		return nil
	}
	value := binary.BigEndian.Uint32(avp.Data)
	return &value
}

func (avp Avp) ToUint64() *uint64 {
	if avp.Data == nil {
		return nil
	}
	value := binary.BigEndian.Uint64(avp.Data)
	return &value
}

func (avp Avp) ToFloat32() *float32 {
	if avp.Data == nil {
		return nil
	}
	bits := binary.BigEndian.Uint32(avp.Data)
	value := math.Float32frombits(bits)
	return &value
}

func (avp Avp) ToFloat64() *float64 {
	if avp.Data == nil {
		return nil
	}
	bits := binary.BigEndian.Uint64(avp.Data)
	value := math.Float64frombits(bits)
	return &value
}

func (avp Avp) ToNetIP() *net.IP {
	if avp.Data == nil {
		return nil
	}
	if avp.Data[1] == 1 {
		value := net.IP(avp.Data[2:6])
		return &value
	} else {
		value := net.IP(avp.Data[2:18])
		return &value
	}
}

func (avp Avp) ToTime() *time.Time {
	if avp.Data == nil {
		return nil
	}
	timestamp := int64(binary.BigEndian.Uint32(avp.Data))
	value := time.Unix(timestamp-2208988800, 0)
	return &value
}

func (avp Avp) ToGroup() Avps {
	return ReadAvps(avp.Data)
}

func ReadAvps(bytes []byte) Avps {
	if bytes == nil {
		return nil
	}
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

func readUInt24(bytes []byte) uint32 {
	if len(bytes) == 3 {
		bytes = append([]byte{0}, bytes[:]...)
	}
	return binary.BigEndian.Uint32(bytes)
}

func ReadMessage(bytes []byte) *Message {
	if len(bytes) < 20 {
		return nil
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
		Avps:          ReadAvps(bytes[20:]),
	}
	return &message
}
