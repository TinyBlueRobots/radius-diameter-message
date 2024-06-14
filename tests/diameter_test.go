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
	avps := make(diameter.Avps, 0)
	avps = avps.AddUint32(258, 0, mandatoryFlags, 1)
	ipAddress := net.IPv4(100, 98, 179, 174).To4()
	avps = avps.AddNetIP(8, 0, mandatoryFlags, ipAddress)
	message := diameter.NewMessage(1, requestFlags, 265, 1, [4]byte{0, 0, 0, 0}, [4]byte{0, 0, 0, 0}, avps)
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
	assert.Equal(t, []byte{0x0, 0x0, 0x2c}, length)
	assert.Equal(t, byte(0x80), flags)
	assert.Equal(t, []byte{0x0, 0x1, 0x9}, commandCode)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x1}, applicationId)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, hopByHopId)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, endToEndId)
	expectedAvp := []byte{0x0, 0x0, 0x1, 0x02, byte(mandatoryFlags), 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1}
	assert.Equal(t, expectedAvp, actualAvps[:12])
	expectedAvp = []byte{0x0, 0x0, 0x0, 0x08, byte(mandatoryFlags), 0x0, 0x0, 0xc, 0x64, 0x62, 0xb3, 0xae}
	assert.Equal(t, expectedAvp, actualAvps[12:])

	message = *diameter.ReadMessage(bytes)
	avp := message.Avps.Get(258, 0)[0]
	assert.Equal(t, uint32(1), *avp.ToUint32())
	avp = message.Avps.Get(8, 0)[0]
	assert.Equal(t, ipAddress, *avp.ToNetIP())
}

func Test_diameter_read_grouped_avp(t *testing.T) {
	base64Data := "AAADaYAAANQAACivAAADaoAAAMgAACivAAAD+IAAADwAACivAAAD/IAAAA8AACivUy01AAAABBCAAAAQAAAorw7msoAAAAQRgAAAEAAAKK8HoSAAAAAABoAAABAAACivUH2ryQAAAAqAAAANAAAorzUAAAAAAAAeAAAAE2RhdGFjb25uZWN0AAAAAA2AAAAQAAAorzA4MDAAAAASgAAAEQAAKK8yMzQxMAAAAAAAA+yAAAATAAAor2RlZmF1bHQAAAAAFoAAABQAACivAAAAAAAAAAA="
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	for _, avp := range diameter.ReadAvps(decodedData) {
		if avp.Code == 873 {
			for _, avp := range avp.ToGroup() {
				for _, avp := range avp.ToGroup() {
					if avp.Code == 30 {
						assert.Equal(t, "dataconnect", *avp.ToString())
						return
					}
				}
			}
		}
	}
	t.Fatal("Grouped AVP not found")
}

func Test_diameter_write_grouped_avp(t *testing.T) {
	avps := make(diameter.Avps, 0)
	group := make(diameter.Avps, 0)
	group = group.AddUint32(432, 0, 0, 1)
	avps = avps.AddGroup(456, 0, 0, group)
	message := diameter.NewMessage(1, 0, 265, 1, [4]byte{0, 0, 0, 0}, [4]byte{0, 0, 0, 0}, avps)
	bytes := message.ToBytes()
	message = *diameter.ReadMessage(bytes)
	for _, avp := range message.Avps {
		if avp.Code == 456 {
			for _, avp := range avp.ToGroup() {
				if avp.Code == 432 {
					assert.Equal(t, uint32(1), *avp.ToUint32())
					return
				}
			}
		}
	}
	t.Fatal("Grouped AVP not found")
}

func Test_diameter_timestamp(t *testing.T) {
	base64Data := "AAAANwAAAAzp72Zd"
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	avp := diameter.ReadAvps(decodedData).Get(55, 0)[0]
	expected := time.Time(time.Date(2024, time.May, 15, 17, 50, 37, 0, time.Local))
	assert.Equal(t, expected, *avp.ToTime())
}

func Test_diameter_bytes(t *testing.T) {
	avp := diameter.NewAvp(1, 0, 0, []byte{0x0, 0x0, 0x0, 0x1})
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1}, avp.ToBytes())
}
