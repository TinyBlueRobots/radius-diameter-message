package tests

import (
	"encoding/base64"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tinybluerobots/radius-diameter-message/radius"
)

func Test_radius_message(t *testing.T) {
	avps := radius.NewAvps()
	avps = avps.AddString(1, 0, "901280064290558")
	avps = avps.AddString(1, 10415, "901280064290558")
	authenticator := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	message := radius.NewMessage(1, 1, authenticator, avps...)
	bytes := message.ToBytes()
	assert.Equal(t, byte(1), bytes[0])
	assert.Equal(t, byte(1), bytes[1])
	assert.Equal(t, []byte{0x0, 0x3c}, bytes[2:4])
	assert.Equal(t, authenticator[:], bytes[4:20])
	assert.Equal(t, []byte{0x1, 0x11, 0x39, 0x30, 0x31, 0x32, 0x38, 0x30, 0x30, 0x36, 0x34, 0x32, 0x39, 0x30, 0x35, 0x35, 0x38}, bytes[20:37])
	assert.Equal(t, []byte{0x1a, 0x17, 0x0, 0x0, 0x28, 0xaf, 0x1, 0x11, 0x39, 0x30, 0x31, 0x32, 0x38, 0x30, 0x30, 0x36, 0x34, 0x32, 0x39, 0x30, 0x35, 0x35, 0x38}, bytes[37:])

	message = *radius.ReadMessage(bytes)
	avp := message.Avps.GetFirst(1, 0).ToString()
	assert.Equal(t, "901280064290558", *avp)
}

func Test_radius_timestamp(t *testing.T) {
	base64Data := "NwZkpTYl"
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	decodedData = append(make([]byte, 20), decodedData...)
	if err != nil {
		t.Fatal(err)
	}
	message := radius.ReadMessage(decodedData)
	avp := message.Avps.GetFirst(55, 0).ToTime()
	expected := time.Time(time.Date(2023, time.July, 5, 10, 21, 41, 0, time.Local))
	assert.Equal(t, expected, *avp)
}

func Test_read_message_bytes(t *testing.T) {
	base64Data := "BIUC2ECaPuKcYNeTCdoQ+dYmU8ABETkwMTI4MDA2NDI5MDU1OCgGAAAAAgQGLr7D/iAQNDYuMTkwLjE5NS4yNTQGBgAAAAIHBgAAAAcIBgqcBJIeFHNjYW5pYS5mbXB0ZXN0LmN4bh8RODgyMzkwMDY0MjkwNTU4LBYyRTZDRkEwODAwRkYxMTExNzg2MzIWMkU2Q0ZBMDgwMEZGMTExMTc4NjMtBgAAAAE9BgAAABI3BmSlNiUqBgABIbIrBgAAs5gvBgAABEgwBgAAAzU0BgAAAAA1BgAAAAAxBgAAAAEuBgAADPIaFwAAKK8BETkwMTI4MDA2NDI5MDU1OBoMAAAorwIGARF4YxoMAAAorwMGAAAAABoMAAAorwQGLmz6CBofAAAorwUZMDgtNkMwOTAwMDAyNzEwMDAwMTg2QTAaDAAAKK8GBlAKAoMaDAAAKK8HBi5s+ggaDQAAKK8IBzkwMTI4Gg0AACivCQc5MDEyOBoJAAAorwoDNRoJAAAorwsD/xoJAAAorwwDMBoMAAAorw0GMDAwMBoNAAAorxIHMjA4MDEaGAAAKK8UEjg2ODgwODA0Njk2MTM0MDEaCQAAKK8VAwYaFQAAKK8WD4IC+BDNpwL4EAF+DAwaCgAAKK8XBIABGgkAACivGgMaIUhyZmUAAAAABWAkcu25/wUAAAAAAAAAAAACAFI9Lr7D/gAAAAAAAAAAAQAAAIW6ImLL5r8gLoBu2WLEV4AGAQQAAAAAAAAAIaFyYmUAiyZy7bn/BQAAAAAAAAAAAEMAAAD+PhOetTMDAP7aVXKHIgMAWdsrXsPdHgAQY3gRAVAKAoMAAAAAOTAxMjgAAAIAUj0uvsP+AAAAAAAAAABcCAEAhQK6ImLL5r8gLoBu2WLEV4AGDwAKnASSAAAAAAAAAAA5MDEyODAwNjQyOTA1NTgAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAA="
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	message := radius.ReadMessage(decodedData)
	avp := message.Avps.GetFirst(1, 10415).ToString()
	assert.Equal(t, "901280064290558", *avp)
}

func Test_radius_bytes(t *testing.T) {
	avp := radius.NewAvp(1, 0, []byte{0x0, 0x0, 0x0, 0x1})
	assert.Equal(t, []byte{0x1, 0x6, 0x0, 0x0, 0x0, 0x1}, avp.ToBytes())
}

func Test_radius_nil(t *testing.T) {
	var avps radius.Avps
	avp := avps.GetFirst(1, 0).ToString()
	assert.Nil(t, avp)
}

func Test_radius_string_default(t *testing.T) {
	avps := radius.NewAvps()
	avpString := avps.GetFirst(1, 0).ToStringOrDefault()
	assert.Equal(t, "", avpString)
	avpUint32 := avps.GetFirst(1, 0).ToUint32OrDefault()
	assert.Equal(t, uint32(0), avpUint32)
	avpTime := avps.GetFirst(1, 0).ToTimeOrDefault()
	assert.Equal(t, time.Time{}, avpTime)
	avpNetIP := avps.GetFirst(1, 0).ToNetIPOrDefault()
	var defaultNetIp net.IP
	assert.Equal(t, defaultNetIp, avpNetIP)
	avpData := avps.GetFirst(1, 0).ToData()
	assert.Nil(t, avpData)
}
