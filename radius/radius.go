package radius

import (
	"encoding/binary"
	"net"
	"time"
)

type AttributeType byte
type VendorId uint32
type avpData []byte

type avp struct {
	Type     AttributeType
	length   byte
	vendorId VendorId
	Data     avpData
}

func newAvp(attributeType AttributeType, vendorId VendorId, avpData avpData) avp {
	a := avp{
		Type: attributeType,
		Data: avpData,
	}
	if vendorId == 0 {
		a.length = byte(len(avpData) + 2)
	} else {
		a.vendorId = vendorId
		a.length = byte(len(avpData) + 8)
		a.Data = avpData
	}
	return a
}

func (avp avp) toBytes() []byte {
	bytes := make([]byte, 0)
	if avp.vendorId == 0 {
		bytes = append(bytes, byte(avp.Type))
		bytes = append(bytes, avp.length)
	} else {
		bytes = append(bytes, 26)
		bytes = append(bytes, avp.length)
		buffer := make([]byte, 4)
		binary.BigEndian.PutUint32(buffer, uint32(avp.vendorId))
		bytes = append(bytes, buffer...)
		bytes = append(bytes, byte(avp.Type))
		bytes = append(bytes, byte(len(avp.Data)+2))
	}
	bytes = append(bytes, avp.Data...)
	return bytes
}

type avpId struct {
	attributeType AttributeType
	vendorId      VendorId
}

type Avps map[avpId][]avp

func (avps Avps) Add(attributeType AttributeType, vendorId VendorId, data avpData) {
	avpId := avpId{attributeType, vendorId}
	if avps[avpId] == nil {
		avps[avpId] = make([]avp, 0)
	}
	avps[avpId] = append(avps[avpId], newAvp(attributeType, vendorId, data))
}

type Code uint32

type Message struct {
	Code          Code
	Identifier    byte
	Length        uint16
	Authenticator [16]byte
	Avps          Avps
}

func NewMessage(code Code, identifier byte, authenticator [16]byte, avps Avps) Message {
	length := uint16(20)
	for _, avp := range avps {
		for _, avp := range avp {
			length += uint16(avp.length)
		}
	}
	return Message{
		Code:          code,
		Identifier:    identifier,
		Length:        length,
		Authenticator: authenticator,
		Avps:          avps,
	}
}

func (message Message) ToBytes() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, byte(message.Code))
	bytes = append(bytes, message.Identifier)
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, message.Length)
	bytes = append(bytes, buffer...)
	bytes = append(bytes, message.Authenticator[:]...)
	for _, avp := range message.Avps {
		for _, avp := range avp {
			bytes = append(bytes, avp.toBytes()...)
		}
	}
	return bytes
}

func (avps Avps) Get(attributeType AttributeType, vendorId VendorId) []avp {
	return avps[avpId{attributeType, vendorId}]
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
	value := time.Unix(timestamp, 0)
	return &value
}

func ReadAvps(bytes []byte) Avps {
	offset := 0
	avps := make(Avps)
	for offset < len(bytes) {
		attributeType := AttributeType(bytes[offset])
		length := bytes[offset+1]
		avpId := avpId{attributeType, 0}
		var avpData avpData
		if attributeType == 26 {
			avpId.vendorId = VendorId(binary.BigEndian.Uint32(bytes[offset+2 : offset+6]))
			attributeType = AttributeType(bytes[offset+6])
			avpId.attributeType = attributeType
			avpLength := bytes[offset+7]
			avpData = bytes[offset+8 : offset+6+int(avpLength)]
		} else {
			avpLength := bytes[offset+1]
			avpData = bytes[offset+2 : offset+int(avpLength)]
		}
		if avps[avpId] == nil {
			avps[avpId] = make([]avp, 0)
		}
		avps[avpId] = append(avps[avpId], newAvp(attributeType, avpId.vendorId, avpData))
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
		Length:        binary.BigEndian.Uint16(bytes[2:4]),
		Authenticator: authenticator,
		Avps:          ReadAvps(bytes[20:]),
	}
	return &message
}
