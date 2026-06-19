package vuxml

import (
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/ulikunitz/xz"
)

var vuxmlClient = &http.Client{
	Timeout: 60 * time.Second,
}

type VuxmlEntry struct {
	Vid        string
	Topic      string
	Entry      string
	Packages   []string
	Cves       []string
	Paragraphs []string
}

type vuxmlPackage struct {
	Names []string `xml:"name"`
}

type vuxmlVuln struct {
	Vid         string         `xml:"vid,attr"`
	Topic       string         `xml:"topic"`
	Packages    []vuxmlPackage `xml:"affects>package"`
	Cves        []string       `xml:"references>cvename"`
	Entry       string         `xml:"dates>entry"`
	Description struct {
		Inner string `xml:",innerxml"`
	} `xml:"description"`
}

type vuxmlRoot struct {
	Vulns []vuxmlVuln `xml:"vuln"`
}

func (e *VuxmlEntry) Reference(pkgName string) (ref string, ok bool) {
	entry := e.Entry
	if entry == "" || pkgName == "" {
		return
	}

	year := entry
	if idx := strings.IndexByte(entry, '-'); idx > 0 {
		year = entry[:idx]
	}
	if len(year) != 4 {
		return
	}

	sum := crc32.ChecksumIEEE([]byte(pkgName + entry))
	ref = fmt.Sprintf("FBSD-%s:%08X", year, sum)
	ok = true

	return
}

func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func extractParagraphs(inner string) (paragraphs []string) {
	paragraphs = []string{}

	dec := xml.NewDecoder(strings.NewReader(inner))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity

	depth := 0
	buf := strings.Builder{}

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "p" {
				if depth == 0 {
					buf.Reset()
				}
				depth += 1
			}
		case xml.EndElement:
			if el.Name.Local == "p" && depth > 0 {
				depth -= 1
				if depth == 0 {
					text := collapseSpace(buf.String())
					if text != "" {
						paragraphs = append(paragraphs, text)
					}
				}
			}
		case xml.CharData:
			if depth > 0 {
				buf.Write(el)
			}
		}
	}

	return
}

func ParseVuxml(data []byte) (entries map[string]*VuxmlEntry, err error) {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity

	root := &vuxmlRoot{}
	err = dec.Decode(root)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "advisory: Failed to parse vuxml"),
		}
		return
	}

	entries = map[string]*VuxmlEntry{}

	for i := range root.Vulns {
		vuln := &root.Vulns[i]
		if vuln.Vid == "" {
			continue
		}

		entry := &VuxmlEntry{
			Vid:        vuln.Vid,
			Topic:      collapseSpace(vuln.Topic),
			Entry:      strings.TrimSpace(vuln.Entry),
			Cves:       vuln.Cves,
			Paragraphs: extractParagraphs(vuln.Description.Inner),
		}

		for _, pkg := range vuln.Packages {
			entry.Packages = append(entry.Packages, pkg.Names...)
		}

		entries[vuln.Vid] = entry
	}

	return
}

func Load() (entries map[string]*VuxmlEntry, err error) {
	req, err := http.NewRequest("GET", FreeBsdVuXml, nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "advisory: Failed to create vuxml request"),
		}
		return
	}

	resp, err := vuxmlClient.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "advisory: Failed to request vuxml"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = &errortypes.RequestError{
			errors.Newf(
				"advisory: Bad vuxml status %d", resp.StatusCode),
		}
		return
	}

	xzReader, err := xz.NewReader(resp.Body)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "advisory: Failed to open vuxml xz reader"),
		}
		return
	}

	data, err := io.ReadAll(io.LimitReader(
		xzReader, int64(settings.Telemetry.VuxmlSizeLimit)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "advisory: Failed to read vuxml"),
		}
		return
	}

	entries, err = ParseVuxml(data)
	if err != nil {
		return
	}

	return
}
