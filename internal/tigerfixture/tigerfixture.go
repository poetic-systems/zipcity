// Package tigerfixture trims a TIGER/Line archive down to the records a test
// needs.
//
// The Census Bureau ships one archive per county, per state, or for the whole
// country, and the smallest of them is larger than a test fixture has any
// business being. Keeping a subset needs no shapefile writer: every one of the
// three formats in the archive is a header followed by records that can be
// copied verbatim.
//
//   - .dbf is a fixed-width header and fixed-width records. A kept record is a
//     byte copy and the record count in the header is the only field to patch.
//   - .shp records are length-prefixed and self-contained, so a kept record
//     copies verbatim and only needs renumbering.
//   - .shx is not copied at all. It is the offset and length of each .shp
//     record, so it is regenerated from the records that survive.
//
// The bounding box in a .shp header is left as the source wrote it. It stays a
// true bound of the records that remain, just no longer the tightest one.
//
// The ISO metadata alongside the data describes the full file, so a subset
// drops it rather than shipping a description that no longer fits.
package tigerfixture

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-shapefile"
)

const (
	dbfHeaderOffset = 32  // where the field descriptors start
	shpHeaderSize   = 100 // shared by .shp and .shx
	shxRecordSize   = 8
)

// Keep reports whether the record with the given attributes belongs in the
// subset. Records are offered in file order.
type Keep func(fields map[string]any) bool

// Subset writes to dst an archive holding only the records of src that keep
// accepts, and reports how many it kept.
func Subset(src, dst string, keep Keep) (int, error) {
	members, err := readMembers(src)
	if err != nil {
		return 0, err
	}

	var kept []int
	at := 0
	if err := Scan(src, func(fields map[string]any) {
		if keep(fields) {
			kept = append(kept, at)
		}
		at++
	}); err != nil {
		return 0, err
	}

	if err := writeSubset(dst, members, kept); err != nil {
		return 0, err
	}
	return len(kept), nil
}

// Scan calls visit with the attributes of each record of src, in file order.
// Choosing what to keep usually means reading one file to learn what another
// joins to, which is a read and not a subset.
func Scan(src string, visit func(fields map[string]any)) error {
	shp, err := shapefile.ReadZipFile(src, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", src, err)
	}
	for i := range shp.NumRecords() {
		fields, _ := shp.Record(i)
		visit(fields)
	}
	return nil
}

// member is one file inside the archive, held whole. TIGER archives are a
// handful of files and the largest is read into memory by the reader anyway.
type member struct {
	name string
	data []byte
}

func readMembers(src string) ([]member, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src, err)
	}
	defer reader.Close()

	var members []member
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "__MACOSX/") {
			continue
		}
		opened, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", src, file.Name, err)
		}
		data, err := io.ReadAll(opened)
		opened.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", src, file.Name, err)
		}
		members = append(members, member{name: file.Name, data: data})
	}
	return members, nil
}

func writeSubset(dst string, members []member, kept []int) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	archive := zip.NewWriter(out)
	write := func(name string, data []byte) error {
		entry, err := archive.Create(name)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", dst, name, err)
		}
		_, err = entry.Write(data)
		return err
	}

	for _, m := range members {
		lower := strings.ToLower(m.name)
		switch {
		case strings.HasSuffix(lower, ".xml"):
			continue
		case strings.HasSuffix(lower, ".shx"):
			continue // regenerated alongside the .shp
		case strings.HasSuffix(lower, ".dbf"):
			data, err := subsetDBF(m.data, kept)
			if err != nil {
				return fmt.Errorf("%s: %s: %w", dst, m.name, err)
			}
			if err := write(m.name, data); err != nil {
				return err
			}
		case strings.HasSuffix(lower, ".shp"):
			shp, shx, err := subsetSHP(m.data, kept)
			if err != nil {
				return fmt.Errorf("%s: %s: %w", dst, m.name, err)
			}
			if err := write(m.name, shp); err != nil {
				return err
			}
			if err := write(m.name[:len(m.name)-len(".shp")]+".shx", shx); err != nil {
				return err
			}
		default:
			if err := write(m.name, m.data); err != nil {
				return err
			}
		}
	}
	return archive.Close()
}

// subsetDBF copies the header and the kept records, patching the record count.
func subsetDBF(data []byte, kept []int) ([]byte, error) {
	if len(data) < dbfHeaderOffset {
		return nil, fmt.Errorf("dbf is %d bytes, shorter than its header", len(data))
	}
	headerSize := int(binary.LittleEndian.Uint16(data[8:10]))
	recordSize := int(binary.LittleEndian.Uint16(data[10:12]))
	if headerSize < dbfHeaderOffset || recordSize == 0 {
		return nil, fmt.Errorf("dbf header claims a %d byte header and %d byte records", headerSize, recordSize)
	}

	out := make([]byte, 0, headerSize+len(kept)*recordSize+1)
	out = append(out, data[:headerSize]...)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(kept)))

	for _, i := range kept {
		start := headerSize + i*recordSize
		if start+recordSize > len(data) {
			return nil, fmt.Errorf("dbf record %d runs past the end of the file", i)
		}
		out = append(out, data[start:start+recordSize]...)
	}
	return append(out, 0x1a), nil
}

// subsetSHP copies the kept shape records, renumbering them, and builds the
// .shx that indexes what it wrote.
func subsetSHP(data []byte, kept []int) (shp, shx []byte, err error) {
	offsets, err := shpRecordOffsets(data)
	if err != nil {
		return nil, nil, err
	}

	shp = append(shp, data[:shpHeaderSize]...)
	shx = append(shx, data[:shpHeaderSize]...)

	for n, i := range kept {
		if i >= len(offsets) {
			return nil, nil, fmt.Errorf("shp has %d records, so record %d is not in it", len(offsets), i)
		}
		start, end := offsets[i], offsets[i]+8+contentLength(data, offsets[i])

		shx = binary.BigEndian.AppendUint32(shx, uint32(len(shp)/2))
		shx = binary.BigEndian.AppendUint32(shx, uint32((end-start-8)/2))

		record := append([]byte(nil), data[start:end]...)
		binary.BigEndian.PutUint32(record[:4], uint32(n+1))
		shp = append(shp, record...)
	}

	binary.BigEndian.PutUint32(shp[24:28], uint32(len(shp)/2))
	binary.BigEndian.PutUint32(shx[24:28], uint32(len(shx)/2))
	return shp, shx, nil
}

// shpRecordOffsets walks the record headers, which is the only way to find the
// records: a .shp file has no index of its own.
func shpRecordOffsets(data []byte) ([]int, error) {
	var offsets []int
	for at := shpHeaderSize; at < len(data); {
		if at+8 > len(data) {
			return nil, fmt.Errorf("shp record header at %d runs past the end of the file", at)
		}
		offsets = append(offsets, at)
		at += 8 + contentLength(data, at)
	}
	return offsets, nil
}

func contentLength(data []byte, recordHeader int) int {
	return 2 * int(binary.BigEndian.Uint32(data[recordHeader+4:recordHeader+8]))
}
