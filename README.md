# zipcity

Encoding US zip code to city mappings efficiently for use in address parsers

## Copyright and License

Copyright 2026 Poetic Systems

Unless otherwise specified all code and related artifacts in this repository are made available under the Apache 2 License. See the [license](./LICENSE.md) for details.

### Data

Data used to build this repository is derived from freely available sources and is treated as a software dependency, licensed according to it's original terms.

Zip / Postal Code data is from [Geonames](https://download.geonames.org/export/zip/readme.txt)
and liscensed under a Creative Commons Attribution 4.0 License.

[US Census Bureau Tiger Line/Shapefiles](https://www.census.gov/geographies/mapping-files/time-series/geo/tiger-line-file.html) are available without cost because [the U.S. government releases these publications into the public domain](https://www.census.gov/about/policies/open-gov.html).

## Pupose

### Address Parsing is Data Dependant

When building an address parser - such as [go-projectusat](https://github.com/PortobelloAuth/go-projectusat) - it swiftly becomes apparent that it is valuable to have a database of real postal codes, municipalities, and street names. You don't have to look far to find commercial address parsing, normalization, and verification services. It is significantly more difficult to find open source libraries that can do the same. Address parsing is hard (even when you're trying to follow a standard like [USPS Pub 28](https://pe.usps.com/text/pub28/welcome.htm) or [Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf).)

One of the more difficult challenges in address parsing is dealing with ambiguity in the address itself. Addresses are not a [regular language](https://en.wikipedia.org/wiki/Regular_language) even within the reduced scope of US addresses. In many cases determining how an address should be parsed requires knowing what values are valid for specific elements of the address. For instance:

    3253 W 9200 S WEST JORDAN UT 84088

Without knowing anything about cities in Utah, could be parsed as:

```go
  PrimaryNumber:       "3253",
  Predirectional:      "W",
  StreetName:          "9200",
  StreetSuffix:        "",
  Postdirectional:     "S",
  SecondaryDesignator: "",
  SecondaryNumber:     "",
  City:                "WEST JORDAN",
  Region:              "UT",
  Postal:              "84088",
  Country:             "",
```

Or:

```go
  PrimaryNumber:       "3253",
  Predirectional:      "W",
  StreetName:          "9200",
  StreetSuffix:        "",
  Postdirectional:     "SW",
  SecondaryDesignator: "",
  SecondaryNumber:     "",
  City:                "JORDAN",
  Region:              "UT",
  Postal:              "84088",
  Country:             "",
```

Or even:

```go
  PrimaryNumber:       "3253",
  Predirectional:      "W",
  StreetName:          "9200",
  StreetSuffix:        "",
  Postdirectional:     "",
  SecondaryDesignator: "",
  SecondaryNumber:     "",
  City:                "SOUTHWEST JORDAN",
  Region:              "UT",
  Postal:              "84088",
  Country:             "",
```

A similar problem occurs with addresses like

    123 E ST IDAHO FALLS ID 83402

where it is not clear whether "E ST" means "EAST ST" or "E ST". Similar problems can occur in cites like "ST PAUL, MN", where "ST" means "Saint" as part of the city name, not "Street" as the street suffix. Determining what is correct is all data dependent.

### Addresses Can Be Associated With People

People live at addresses. In the Project US@ case, those people are patients. Calling a service to validate an address is therefore sharing personal health information with that service in a way that could reveal sensitive data such as where a patient is recieving treatment. Nearly every entity in healthcare will need to validate a patient's address at some point - potentially quite frequently as part of records exchange.

### Calling a Service is Expensive

In order to maintain functionality, services are required to throttle access and, generally, charge for different rate limts. By their very nature, services cost network bandwidth and latency to call. Caching results is common, but presents it's own set of privacy and security concerns.

### This Library's Approach

If we can sufficiently compress the needed data it can be distributed with the parser and kept in memory. This eliminates both network latency and service cost concerns. It also protects patient privacy because there is no additional network request to log.

(To be tested:) By converting indicies of street, city, region, and zip code data in to bloom filters we can significantly reduce the size of the data needed to determine a valid street name or city name for a given zip code or city-region tuple. This library, as a build process, processes its input data in to bloom filters which are then serialized in to a form that can be efficiently committed to and distributed with the library.
