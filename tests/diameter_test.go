package tests

import (
	"encoding/base64"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tinybluerobots/radius-diameter-message/diameter"
)

const (
	mandatoryFlags diameter.Flags = 0x40
	requestFlags   diameter.Flags = 0x80
)

func Test_diameter_message(t *testing.T) {
	avps := diameter.NewAvps()
	avps = avps.AddUint32(258, mandatoryFlags, 0, 1)
	ipAddress := net.IPv4(100, 98, 179, 174)
	avps = avps.AddNetIP(257, mandatoryFlags, 0, ipAddress)
	message := diameter.NewMessage(1, requestFlags, 265, 1, [4]byte{0, 0, 0, 0}, [4]byte{0, 0, 0, 0}, avps...)
	bytes := message.ToBytes()
	version := bytes[0]
	length := bytes[1:4]
	flags := bytes[4]
	commandCode := bytes[5:8]
	applicationId := bytes[8:12]
	hopByHopId := bytes[12:16]
	endToEndId := bytes[16:20]
	actualAvps := bytes[20:]
	assert.Equal(t, byte(1), version)
	assert.Equal(t, []byte{0x0, 0x0, 0x30}, length)
	assert.Equal(t, byte(0x80), flags)
	assert.Equal(t, []byte{0x0, 0x1, 0x9}, commandCode)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x1}, applicationId)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, hopByHopId)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, endToEndId)
	expectedAvp := []byte{0x0, 0x0, 0x1, 0x02, byte(mandatoryFlags), 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1}
	assert.Equal(t, expectedAvp, actualAvps[:12])
	expectedAvp = []byte{0x0, 0x0, 0x1, 0x1, byte(mandatoryFlags), 0x0, 0x0, 0xe, 0x0, 0x1, 0x64, 0x62, 0xb3, 0xae, 0x0, 0x0}
	assert.Equal(t, expectedAvp, actualAvps[12:])

	message = *diameter.ReadMessage(bytes)
	avp := message.Avps.GetFirst(258, 0)
	assert.Equal(t, uint32(1), *avp.ToUint32())
	avp = message.Avps.GetFirst(257, 0)
	assert.Equal(t, ipAddress.To4(), *avp.ToNetIP())
}

func Test_diameter_read_grouped_avp(t *testing.T) {
	base64Data := "AAADaYAAANQAACivAAADaoAAAMgAACivAAAD+IAAADwAACivAAAD/IAAAA8AACivUy01AAAABBCAAAAQAAAorw7msoAAAAQRgAAAEAAAKK8HoSAAAAAABoAAABAAACivUH2ryQAAAAqAAAANAAAorzUAAAAAAAAeAAAAE2RhdGFjb25uZWN0AAAAAA2AAAAQAAAorzA4MDAAAAASgAAAEQAAKK8yMzQxMAAAAAAAA+yAAAATAAAor2RlZmF1bHQAAAAAFoAAABQAACivAAAAAAAAAAA="
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	messageData := make([]byte, 20+len(decodedData))
	copy(messageData[20:], decodedData)
	message := *diameter.ReadMessage(messageData)
	apn := message.Avps.GetFirst(873, 10415).ToGroup().GetFirst(874, 10415).ToGroup().GetFirst(30, 0).ToString()
	assert.Equal(t, "dataconnect", *apn)
}

func Test_diameter_nil(t *testing.T) {
	var avps diameter.Avps
	value := avps.GetFirst(1, 0).ToGroup().GetFirst(1, 0).ToString()
	assert.Nil(t, value)
}

func Test_diameter_string_default(t *testing.T) {
	avps := diameter.NewAvps()
	avpString := avps.GetFirst(1, 0).ToStringOrDefault()
	assert.Equal(t, "", avpString)
	avpUint32 := avps.GetFirst(1, 0).ToUint32OrDefault()
	assert.Equal(t, uint32(0), avpUint32)
	avpUint64 := avps.GetFirst(1, 0).ToUint64OrDefault()
	assert.Equal(t, uint64(0), avpUint64)
	avpTime := avps.GetFirst(1, 0).ToTimeOrDefault()
	assert.Equal(t, time.Time{}, avpTime)
	avpNetIP := avps.GetFirst(1, 0).ToNetIPOrDefault()
	var defaultNetIp net.IP
	assert.Equal(t, defaultNetIp, avpNetIP)
	avpData := avps.GetFirst(1, 0).ToData()
	assert.Nil(t, avpData)
}

func Test_diameter_write_grouped_avp(t *testing.T) {
	group := diameter.NewAvps()
	group = group.AddUint32(432, 0, 0, 1)
	avps := diameter.NewAvps()
	avps = avps.AddGroup(456, 0, 0, group...)
	message := diameter.NewMessage(1, 0, 265, 1, [4]byte{0, 0, 0, 0}, [4]byte{0, 0, 0, 0}, avps...)
	bytes := message.ToBytes()
	message = *diameter.ReadMessage(bytes)
	avp := message.Avps.GetFirst(456, 0).ToGroup().GetFirst(432, 0)
	assert.Equal(t, uint32(1), *avp.ToUint32())
}

func Test_diameter_write_grouped_avp_with_spread(t *testing.T) {
	avps := diameter.NewAvpGroup(456, 0, 0, diameter.NewAvpUint32(432, 0, 0, 1))
	message := diameter.NewMessage(1, 0, 265, 1, [4]byte{0, 0, 0, 0}, [4]byte{0, 0, 0, 0}, avps)
	bytes := message.ToBytes()
	message = *diameter.ReadMessage(bytes)
	avp := message.Avps.GetFirst(456, 0).ToGroup().GetFirst(432, 0)
	assert.Equal(t, uint32(1), *avp.ToUint32())
}

func Test_diameter_timestamp(t *testing.T) {
	base64Data := "AAAANwAAAAzp72Zd"
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	messageData := make([]byte, 20+len(decodedData))
	copy(messageData[20:], decodedData)
	message := *diameter.ReadMessage(messageData)
	avp := message.Avps.GetFirst(55, 0)
	expected := time.Time(time.Date(2024, time.May, 15, 17, 50, 37, 0, time.Local))
	assert.Equal(t, expected, *avp.ToTime())
}

func Test_diameter_bytes(t *testing.T) {
	avp := diameter.NewAvp(1, 0, 0, []byte{0x0, 0x0, 0x0, 0x1})
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1}, avp.ToBytes())
}

func Test_diameter_vendor_avp(t *testing.T) {
	base64Data := "AAADZcAAABAAACivBPc8Zg=="
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	avp := diameter.NewAvpUint32(869, 0xc0, 10415, 83311718)
	assert.Equal(t, decodedData, avp.ToBytes())
}
