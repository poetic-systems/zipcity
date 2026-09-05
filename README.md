# zipcity

Encoding US zip code to city mappings efficiently for use in address parsers

## Copyright and License

Copyright 2026 Poetic Systems

Unless otherwise specified all code and related artifacts in this repository are
made available under the Apache 2 License. See the [license](./LICENSE.md) for
details.

### Data

Data used to build this repository is derived from freely available sources and
is treated as a software dependency, licensed according to it's original terms.

Zip / Postal Code data is from
[Geonames](https://download.geonames.org/export/zip/readme.txt) and liscensed
under a Creative Commons Attribution 4.0 License.

[US Census Bureau Tiger Line/Shapefiles](https://www.census.gov/geographies/mapping-files/time-series/geo/tiger-line-file.html)
are available without cost because
[the U.S. government releases these publications into the public domain](https://www.census.gov/about/policies/open-gov.html).

## Purpose

### Address Parsing is Data Dependant

When building an address parser - such as
[go-projectusat](https://github.com/PortobelloAuth/go-projectusat) - it swiftly
becomes apparent that it is valuable to have a database of real postal codes,
municipalities, and street names. You don't have to look far to find commercial
address parsing, normalization, and verification services. It is significantly
more difficult to find open source libraries that can do the same. Address
parsing is hard (even when you're trying to follow a standard like
[USPS Pub 28](https://pe.usps.com/text/pub28/welcome.htm) or
[Project US@](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf).)

One of the more difficult challenges in address parsing is dealing with
ambiguity in the address itself. Addresses are not a
[regular language](https://en.wikipedia.org/wiki/Regular_language) even within
the reduced scope of US addresses. In many cases determining how an address
should be parsed requires knowing what values are valid for specific elements of
the address. For instance:

```text
3253 W 9200 S WEST JORDAN UT 84088
```

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

```text
123 E ST IDAHO FALLS ID 83402
```

where it is not clear whether "E ST" means "EAST ST" or the alphabetical "E ST".
Similar problems can occur in cites like "ST PAUL, MN", where "ST" means "Saint"
as part of the city name, not "Street" as the street suffix. In each case,
determining the correct interpretation of address components is data dependent.

### Addresses Can Be Associated With People

People live at addresses. In the Project US@ case, those people are patients.
Calling a service to validate an address is therefore sharing personal health
information with that service in a way that could reveal sensitive data such as
where a patient is recieving treatment. Nearly every entity in healthcare will
need to validate a patient's address at some point - potentially quite
frequently as part of records exchange.

### Calling a Service is Expensive

In order to maintain functionality, services are required to throttle access
and, generally, charge for different rate limts. By their very nature, services
cost network bandwidth and latency to call. Caching results is common, but
presents it's own set of privacy and security concerns.

### This Library's Approach

If we can sufficiently compress the needed data it can be distributed with the
parser and kept in memory. This eliminates both network latency and service cost
concerns. It also protects patient privacy because there is no additional
network request to log.

By converting indicies of street, city, region, and zip code data in to bloom
filters we can significantly reduce the size of the data needed to determine a
valid street name or city name for a given zip code or city-region tuple. This
library, as a pre-build code generation process, processes its input data in to
bloom filters which are then serialized in to relatively small binary files that
can be efficiently committed to and distributed with the library. These files
are embedded via `go:embed` during the build process so they are both stored and
distributed as efficiently as possible and readily available for use by the
library immediately.

## Building the Data

Both of the data-producing steps in this repository are Go programs carrying
`//go:build ignore`, so they never enter the build graph of the released library
and are run explicitly:

```sh
go run internal/bloomgenerator/bloomgenerator.go
go run internal/tigerfixture/build/generate.go
```

Run them from the module root. Neither is needed to build or use the library —
the compiled filters and the test fixtures are both committed.

### The bloom filters

`internal/bloomgenerator/bloomgenerator.go` downloads what it needs and writes
`generated/compiled_filter/`, which the library embeds.

- **Inputs**, cached on disk and re-used across runs: the Census Bureau TIGER
  files under `./data/us_census_tiger/`, and the GeoNames postal code exports
  under `./data/geonames/`. The first run fetches several gigabytes of TIGER
  files; later runs only fetch what is missing.
- **A required file that will not load stops the run.** Every file the Census
  Bureau's own index lists must be present and readable, and an archive holding
  no files counts as unreadable — a request is sometimes answered with a 200 and
  an HTML error page, which would otherwise read as a county with no records.
  Unreadable cached files are deleted so the next run fetches them again, and
  all the failures are reported together. A file the Census Bureau does not
  publish never reaches that list, so it is not a failure.

### The test fixtures

`internal/tigerfixture/build/generate.go` writes the TIGER slice under
`internal/ustigerline/testdata/us_census_tiger/` that `ustigerline`'s tests
read. It builds one connected slice of a single county — faces, then the edges
they bound, then the names and address ranges for those edges — because the
readers exist to follow the joins between the files, and a fixture that breaks
a join tests nothing.

It reads from the same `./data/us_census_tiger/` cache and downloads nothing, so
the county's files have to be there already.

## Current Known Issues

Because the current version of this library is built from the US Census Bureau's
TIGER files, some street names don't appear in the current dataset. There are
two known primary causes:

1. Several TIGER files on the Census Bureau's FTP site return an error page with
   a 200 status code instead of the listed TIGER file. These files are:
   - <https://www2.census.gov/geo/tiger/TIGER2025/EDGES/tl_2025_48239_edges.zip>
   - <https://www2.census.gov/geo/tiger/TIGER2025/ADDR/tl_2025_13193_addr.zip>
   - <https://www2.census.gov/geo/tiger/TIGER2025/FACES/tl_2025_21061_faces.zip>
   - <https://www2.census.gov/geo/tiger/TIGER2025/FACES/tl_2025_21089_faces.zip>
   - <https://www2.census.gov/geo/tiger/TIGER2025/FACES/tl_2025_13273_faces.zip>
   - <https://www2.census.gov/geo/tiger/TIGER2025/FACES/tl_2025_42065_faces.zip>
     Additionally, there are no address range files for the Marshal Islands and
     the Northern Marianas Islands.
2. TIGER spells a street name in its own vocabulary, which is neither the
   caller's nor the standard's. Puerto Rico street names carry their type at the
   front, and TIGER abbreviates some of those types and not others — `CLL` for
   Calle, `CAM` for Camino, `QBDA` for Quebrada, while writing `EXPRESO`,
   `AUTOPISTA` and `CALLEJÓN` out in full, diacritics included. Project US@ page
   26 forbids abbreviating a street name, so a conforming caller sends `CALLE`
   and the data holds `CLL`.

   The leading type is now aliased at lookup time, so both spellings find the
   street. The rest of the name is not aliased: `CLL LOIZA` and `CLL LOÍZA` are
   separate TIGER records and therefore separate keys, and nothing derives one
   from the other. Folding diacritics in the name half needs the filters rebuilt
   with folded keys — see
   [#1](https://github.com/poetic-systems/zipcity/issues/1).
