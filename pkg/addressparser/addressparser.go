package addressparser

import (
	"fmt"

	"github.com/poetic-systems/zipcity/pkg/address"
)

/*

 */

type ZipCityAddressParser struct{}

func New() *ZipCityAddressParser {
	return &ZipCityAddressParser{}
}

func (z *ZipCityAddressParser) Parse(input string) (*address.ProjectUSAddress, error) {
	return nil, fmt.Errorf("Not implemented")
}
