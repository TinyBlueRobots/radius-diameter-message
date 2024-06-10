package tests

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tinybluerobots/radius-diameter-message/radius"
)

func Test_radius_message(t *testing.T) {
	avps := make(radius.Avps)
	avps.Add(1, 0, []byte("901280064290558"))
	avps.Add(1, 10415, []byte("901280064290558"))
	authenticator := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	message := radius.NewMessage(1, 1, authenticator, avps)
	bytes := message.ToBytes()
	assert.Equal(t, byte(1), bytes[0])
	assert.Equal(t, byte(1), bytes[1])
	assert.Equal(t, []byte{0x0, 0x3c}, bytes[2:4])
	assert.Equal(t, authenticator[:], bytes[4:20])
	assert.Equal(t, []byte{0x1, 0x11, 0x39, 0x30, 0x31, 0x32, 0x38, 0x30, 0x30, 0x36, 0x34, 0x32, 0x39, 0x30, 0x35, 0x35, 0x38}, bytes[20:37])
	assert.Equal(t, []byte{0x1a, 0x17, 0x00, 0x00, 0x28, 0xaf, 0x01, 0x11, 0x39, 0x30, 0x31, 0x32, 0x38, 0x30, 0x30, 0x36, 0x34, 0x32, 0x39, 0x30, 0x35, 0x35, 0x38}, bytes[37:])

	message = *radius.ReadMessage(bytes)
	avp := message.Avps.Get(1, 0)[0]
	assert.Equal(t, "901280064290558", *avp.ToString())
}

func Test_radius_timestamp(t *testing.T) {
	base64Data := "NwZkpTYl"
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		t.Fatal(err)
	}
	avp := radius.ReadAvps(decodedData).Get(55, 0)[0]
	expected := time.Time(time.Date(2023, time.July, 5, 10, 21, 41, 0, time.Local))
	assert.Equal(t, expected, *avp.ToTime())
}
