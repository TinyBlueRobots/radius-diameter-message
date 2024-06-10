# Radius Diameter Message
This is *very* simple Radius and Diameter message reader and writer. It will read `[]bytes` into a `Message` and back again. 

`Avps` are just a map, keyed on AVP Code + Vendor Id, with some helper functions to convert the data into usable types.  

There are no dictionaries, so bring your own, but there are types provided to help create them:  
e.g.

```
const (
	TGPP  VendorId = 10415
	Nokia VendorId = 94
	Cisco VendorId = 9
)
```
