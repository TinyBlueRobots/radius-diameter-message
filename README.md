# Radius Diameter Message

This library provides a simple interface for reading and writing Radius and Diameter messages. It allows for the conversion of `[]bytes` into a `Message` structure, and supports the construction of Messages and Attribute-Value Pairs (AVPs).

## Dictionary types
To keep the library small there are no generated AVP dictionaries included, but types are provided to create your own:

```
type ApplicationId uint32
type Code uint32
type CommandCode uint32
type Flags byte
type VendorId uint32
```

## Usage examples
Examples are for DIAMETER; the RADIUS API is similar but smaller
### Create then read a diameter message
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

### AVP functions
All data types are supported, for brevity the examples will use strings

Create an AVP:  
`avp := diameter.NewAvpString(100, 0, 0x0, "foo")`  

Create and add an AVP to a slice:  
`avps = avps.AddString(100, 0, 0x0, "foo")`  

Read a single AVP of a type:  
`avp := avps.GetFirst(100, 0)`  

Read all the AVPs of a type:  
`filteredAvps := avps.Get(100, 0)`  

Convert an AVP into a grouped AVP:  
`avp := avps.GetFirst(100, 0).ToGroup()`  

Read an AVP value into a pointer:  
`value := avp.ToString()`  

Read an AVP value or use the default if it's nil:  
`value := avp.ToStringOrDefault()`  

Chain these together to read into deeply grouped AVPs:  
`avp := avps.GetFirst(873, 10415).ToGroup().GetFirst(874, 10415).ToGroup().GetFirst(30, 0).ToString()`