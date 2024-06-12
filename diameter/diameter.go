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

type avp struct {
	Code     Code
	Flags    Flags
	length   uint32
	VendorId VendorId
	Data     avpData
	padding  uint32
}

func newAvp(code Code, flags Flags, vendorId VendorId, avpData avpData) avp {
	padding := uint32(len(avpData) % 4)
	if padding != 0 {
		padding = 4 - padding
	}
	length := uint32(len(avpData) + 8)
	if vendorId != 0 {
		length += 4
	}
	return avp{
		Code:     code,
		Flags:    flags,
		length:   length,
		VendorId: vendorId,
		Data:     avpData,
		padding:  padding,
	}
}

func (avp avp) toBytes() []byte {
	bytes := make([]byte, avp.length+avp.padding)
	binary.BigEndian.PutUint32(bytes, uint32(avp.Code))
	bytes[4] = byte(avp.Flags)
	copy(bytes[5:8], writeUInt24(avp.length))
	copy(bytes[8:], avp.Data)
	return bytes
}

type Avps []avp

func (avps Avps) Add(code Code, vendorId VendorId, flags Flags, data avpData) Avps {
	return append(avps, newAvp(code, flags, vendorId, data))
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
	length        uint32
	Flags         Flags
	CommandCode   CommandCode
	ApplicationId ApplicationId
	HopByHopId    [4]byte
	EndToEndId    [4]byte
	Avps          Avps
}

func NewMessage(version byte, flags Flags, commandCode CommandCode, applicationId ApplicationId, hopByHopId [4]byte, endToEndId [4]byte, avps Avps) Message {
	length := uint32(20)
	for _, avp := range avps {
		length += avp.length + avp.padding
	}
	return Message{
		Version:       version,
		length:        length,
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
	bytes = append(bytes, writeUInt24(message.length)...)
	bytes = append(bytes, byte(message.Flags))
	bytes = append(bytes, message.CommandCode.toBytes()...)
	bytes = append(bytes, message.ApplicationId.toBytes()...)
	buffer := make([]byte, 4)
	copy(buffer, message.HopByHopId[:])
	bytes = append(bytes, buffer...)
	copy(buffer, message.EndToEndId[:])
	bytes = append(bytes, buffer...)
	for _, avp := range message.Avps {
		bytes = append(bytes, avp.toBytes()...)
	}
	return bytes
}

func (avps Avps) Get(code Code, vendorId VendorId) Avps {
	filteredAvps := make([]avp, 0)
	for _, avp := range avps {
		if avp.Code == code && avp.VendorId == vendorId {
			filteredAvps = append(filteredAvps, avp)
		}
	}
	return filteredAvps
}

func (avp avp) ToString() *string {
	if avp.Data == nil {
		return nil
	}

	value := string(avp.Data)
	return &value
}

func (avp avp) ToUint32() *uint32 {
	if avp.Data == nil {
		return nil
	}
	value := binary.BigEndian.Uint32(avp.Data)
	return &value
}

func (avp avp) ToFloat32() *float32 {
	if avp.Data == nil {
		return nil
	}
	bits := binary.BigEndian.Uint32(avp.Data)
	value := math.Float32frombits(bits)
	return &value
}

func (avp avp) ToFloat64() *float64 {
	if avp.Data == nil {
		return nil
	}
	bits := binary.BigEndian.Uint64(avp.Data)
	value := math.Float64frombits(bits)
	return &value
}

func (avp avp) ToUint64() *uint64 {
	if avp.Data == nil {
		return nil
	}
	value := binary.BigEndian.Uint64(avp.Data)
	return &value
}

func (avp avp) ToNetIP() *net.IP {
	if avp.Data == nil {
		return nil
	}
	value := net.IP(avp.Data)
	return &value
}

func (avp avp) ToTime() *time.Time {
	if avp.Data == nil {
		return nil
	}
	timestamp := int64(binary.BigEndian.Uint32(avp.Data))
	value := time.Unix(timestamp-2208988800, 0)
	return &value
}

func ReadAvps(bytes []byte) Avps {
	offset := 0
	avps := make(Avps, 0)
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
		avp := newAvp(code, flags, vendorId, avpData)
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
		length:        readUInt24(bytes[1:4]),
		Flags:         Flags(bytes[4]),
		CommandCode:   CommandCode(readUInt24(bytes[5:8])),
		ApplicationId: ApplicationId(binary.BigEndian.Uint32(bytes[8:12])),
		HopByHopId:    hopByHopId,
		EndToEndId:    endToEndId,
		Avps:          ReadAvps(bytes[20:]),
	}
	return &message
}
