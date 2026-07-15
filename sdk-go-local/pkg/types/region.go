package types

// Region identifies a datacenter location for a resource.
//
// Datacenter codes follow the pattern <COUNTRY><CITY>-<City>. The constant
// below covers the value confirmed in the SDK fixtures; the full list is
// available via the API catalog. Any valid region string can be passed
// directly using the typed alias:
//
//	aruba.InRegion(aruba.Region("YOUR-REGION-CODE"))
type Region string

const (
	RegionITBGBergamo Region = "ITBG-Bergamo"
)

// Zone identifies an availability zone within a region for a resource.
//
// Zone codes follow the pattern <REGION>-<INDEX> (e.g., "ITBG-1"). The
// constant below covers the value confirmed in the SDK fixtures. Any valid
// zone string can be passed directly:
//
//	aruba.InZone(aruba.Zone("ITBG-2"))
type Zone string

const (
	ZoneITBG1 Zone = "ITBG-1"
	ZoneITBG2 Zone = "ITBG-2"
	ZoneITBG3 Zone = "ITBG-3"
)
