package cot

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrUIDRequired indicates the CoT event uid field is missing.
	ErrUIDRequired = errors.New("uid is required")
	// ErrTypeRequired indicates the CoT event type field is missing.
	ErrTypeRequired = errors.New("type is required")
	// ErrHowRequired indicates the CoT event how field is missing.
	ErrHowRequired = errors.New("how is required")
	// ErrTimeRequired indicates the CoT event time field is missing.
	ErrTimeRequired = errors.New("time is required")
	// ErrStartRequired indicates the CoT event start field is missing.
	ErrStartRequired = errors.New("start is required")
	// ErrStaleRequired indicates the CoT event stale field is missing.
	ErrStaleRequired = errors.New("stale is required")
	// ErrVersionInvalid indicates an unsupported CoT event version.
	ErrVersionInvalid = errors.New("version must be 2.0")
	// ErrHowInvalid indicates malformed how field.
	ErrHowInvalid = errors.New("how must be a CoT compound hint (for example: m-g)")
	// ErrQoSInvalid indicates malformed qos field.
	ErrQoSInvalid = errors.New("qos must use priority-overtaking-assurance format (for example: 1-r-c)")
	// ErrOpexInvalid indicates malformed opex field.
	ErrOpexInvalid = errors.New("opex must begin with o, e, or s")
)

const (
	// Version2 is the MITRE CoT event schema version most commonly used.
	Version2 = "2.0"
)

// Event models a Cursor on Target (CoT) event payload.
//
// The shape follows the common CoT event envelope:
// - envelope attributes (uid, type, time, start, stale, how)
// - point geometry
// - optional detail metadata.
type Event struct {
	XMLName xml.Name `json:"-" xml:"event"`
	Version string   `json:"version" xml:"version,attr,omitempty"`
	UID     string   `json:"uid" xml:"uid,attr"`
	Type    string   `json:"type" xml:"type,attr"`
	How     string   `json:"how" xml:"how,attr"`
	Access  string   `json:"access,omitempty" xml:"access,attr,omitempty"`
	QoS     string   `json:"qos,omitempty" xml:"qos,attr,omitempty"`
	Opex    string   `json:"opex,omitempty" xml:"opex,attr,omitempty"`

	Time  time.Time `json:"time" xml:"time,attr"`
	Start time.Time `json:"start" xml:"start,attr"`
	Stale time.Time `json:"stale" xml:"stale,attr"`

	Point  Point   `json:"point" xml:"point"`
	Detail *Detail `json:"detail,omitempty" xml:"detail,omitempty"`
}

// Point models CoT geospatial coordinates and optional accuracy values.
type Point struct {
	Lat    float64 `json:"lat" xml:"lat,attr"`
	Lon    float64 `json:"lon" xml:"lon,attr"`
	Hae    float64 `json:"hae" xml:"hae,attr"`
	Ce     float64 `json:"ce" xml:"ce,attr"`
	Le     float64 `json:"le" xml:"le,attr"`
	Alt    float64 `json:"alt,omitempty" xml:"alt,attr,omitempty"`
	Spd    float64 `json:"spd,omitempty" xml:"spd,attr,omitempty"`
	Course float64 `json:"course,omitempty" xml:"course,attr,omitempty"`
}

// Detail captures commonly routed CoT detail fields.
type Detail struct {
	Contact           *Contact           `json:"contact,omitempty" xml:"contact,omitempty"`
	Group             *Group             `json:"group,omitempty" xml:"__group,omitempty"`
	Track             *Track             `json:"track,omitempty" xml:"track,omitempty"`
	Status            *Status            `json:"status,omitempty" xml:"status,omitempty"`
	PrecisionLocation *PrecisionLocation `json:"precision_location,omitempty" xml:"precisionlocation,omitempty"`
	Remarks           string             `json:"remarks,omitempty" xml:"remarks,omitempty"`
}

// Contact identifies an event source or operator.
type Contact struct {
	Callsign string `json:"callsign,omitempty" xml:"callsign,attr,omitempty"`
	Endpoint string `json:"endpoint,omitempty" xml:"endpoint,attr,omitempty"`
}

// Track represents movement or heading metadata.
type Track struct {
	Speed  float64 `json:"speed,omitempty" xml:"speed,attr,omitempty"`
	Course float64 `json:"course,omitempty" xml:"course,attr,omitempty"`
}

// Group represents CoT team/role membership in detail.__group.
type Group struct {
	Name string `json:"name,omitempty" xml:"name,attr,omitempty"`
	Role string `json:"role,omitempty" xml:"role,attr,omitempty"`
}

// PrecisionLocation follows CoT detail.precisionlocation attributes.
type PrecisionLocation struct {
	GeoPointSrc string `json:"geopointsrc,omitempty" xml:"geopointsrc,attr,omitempty"`
	AltSrc      string `json:"altsrc,omitempty" xml:"altsrc,attr,omitempty"`
}

// Status represents lightweight lifecycle or operator state.
type Status struct {
	Battery int    `json:"battery,omitempty" xml:"battery,attr,omitempty"`
	Health  string `json:"health,omitempty" xml:"health,attr,omitempty"`
}

// Validate performs basic field-level checks for routing safety.
func (e Event) Validate() error {
	if e.Version != "" && e.Version != Version2 {
		return ErrVersionInvalid
	}
	if e.UID == "" {
		return ErrUIDRequired
	}
	if e.Type == "" {
		return ErrTypeRequired
	}
	if e.How == "" {
		return ErrHowRequired
	}
	if !isCompound(e.How) {
		return ErrHowInvalid
	}
	if e.Time.IsZero() {
		return ErrTimeRequired
	}
	if e.Start.IsZero() {
		return ErrStartRequired
	}
	if e.Stale.IsZero() {
		return ErrStaleRequired
	}
	if e.Stale.Before(e.Start) {
		return fmt.Errorf("stale must be greater than or equal to start")
	}
	if e.QoS != "" && !isQoS(e.QoS) {
		return ErrQoSInvalid
	}
	if e.Opex != "" && !isOpex(e.Opex) {
		return ErrOpexInvalid
	}
	if e.Point.Lat < -90 || e.Point.Lat > 90 {
		return fmt.Errorf("point.lat out of range: %f", e.Point.Lat)
	}
	if e.Point.Lon < -180 || e.Point.Lon > 180 {
		return fmt.Errorf("point.lon out of range: %f", e.Point.Lon)
	}
	if e.Point.Ce < 0 {
		return fmt.Errorf("point.ce must be non-negative")
	}
	if e.Point.Le < 0 {
		return fmt.Errorf("point.le must be non-negative")
	}

	return nil
}

func isCompound(v string) bool {
	parts := strings.Split(v, "-")
	return len(parts) >= 2 && parts[0] != "" && parts[1] != ""
}

func isQoS(v string) bool {
	parts := strings.Split(v, "-")
	if len(parts) != 3 {
		return false
	}
	if len(parts[0]) != 1 || parts[0][0] < '0' || parts[0][0] > '9' {
		return false
	}

	overtaking := parts[1]
	assurance := parts[2]
	return (overtaking == "r" || overtaking == "f" || overtaking == "i") &&
		(assurance == "g" || assurance == "d" || assurance == "c")
}

func isOpex(v string) bool {
	if v == "" {
		return false
	}
	head := strings.SplitN(v, "-", 2)[0]
	return head == "o" || head == "e" || head == "s"
}
