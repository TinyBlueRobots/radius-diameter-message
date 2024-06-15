# Radius Diameter Message
A simple Radius and Diameter message reader and writer. It will read `[]bytes` into a `Message` and back again, and allow you to construct AVPs. There are helper functions to convert the data into usable types.

There are no dictionaries, so bring your own, but there are types provided to help create them:  
e.g.

```
const (
	TGPP  VendorId = 10415
	Nokia VendorId = 94
	Cisco VendorId = 9
)
```

Create then read a diameter message
```
avps := diameter.NewAvps()
avps = avps.AddUint32(258, 0, mandatoryFlags, 1)
ipAddress := net.IPv4(100, 98, 179, 174).To4()
avps = avps.AddNetIP(8, 0, mandatoryFlags, ipAddress)
message := diameter.NewMessage(1, requestFlags, 265, 1, [4]byte{0, 0, 0, 0}, [4]byte{0, 0, 0, 0}, avps)
bytes := message.ToBytes()

message = *diameter.ReadMessage(bytes)
avp := message.Avps.GetFirst(258, 0)
assert.Equal(t, uint32(1), *avp.ToUint32())
avp = message.Avps.GetFirst(8, 0)
assert.Equal(t, ipAddress, *avp.ToNetIP())
```
