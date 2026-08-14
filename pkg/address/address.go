package address

// ProjectUSAddress captures the address elements that Project US@ and USPS Pub 28
// define as part of an address
type ProjectUSAddress struct {
	BusinessName string // firm / business line (optional)

	// Street line elements
	PrimaryNumber       string
	Predirectional      string
	StreetName          string
	StreetSuffix        string
	Postdirectional     string
	SecondaryDesignator string // APT, STE, ...
	SecondaryNumber     string

	// Last line
	City   string
	Region string // state / province / military "state"
	Postal string // ZIP, ZIP+4, or international postal code

	Country string // optional; often blank for domestic
}
