package cot_test

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cybint/sao/internal/cot"
)

func TestEventValidateSuccess(t *testing.T) {
	now := time.Now().UTC()
	event := cot.Event{
		Version: cot.Version2,
		UID:     "alpha-1",
		Type:    "a-f-G-U-C",
		How:     "m-g",
		Access:  "unrestricted",
		QoS:     "1-r-c",
		Opex:    "o",
		Time:    now,
		Start:   now,
		Stale:   now.Add(5 * time.Minute),
		Point: cot.Point{
			Lat: 34.05,
			Lon: -118.25,
			Hae: 120.0,
			Ce:  10.0,
			Le:  15.0,
		},
	}

	if err := event.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestEventXMLShapeMatchesCoTEnvelope(t *testing.T) {
	now := time.Now().UTC()
	event := cot.Event{
		Version: cot.Version2,
		UID:     "alpha-xml",
		Type:    "a-f-G-U-C",
		How:     "m-g",
		Time:    now,
		Start:   now,
		Stale:   now.Add(3 * time.Minute),
		Point: cot.Point{
			Lat: 1,
			Lon: 2,
			Hae: 3,
			Ce:  10,
			Le:  10,
		},
		Detail: &cot.Detail{
			Group: &cot.Group{
				Name: "Blue",
				Role: "Team Member",
			},
			PrecisionLocation: &cot.PrecisionLocation{
				GeoPointSrc: "GPS",
				AltSrc:      "GPS",
			},
		},
	}

	raw, err := xml.Marshal(event)
	if err != nil {
		t.Fatalf("marshal xml: %v", err)
	}

	output := string(raw)
	if !strings.Contains(output, "<event") {
		t.Fatalf("expected event root element, got %s", output)
	}
	if !strings.Contains(output, "__group") {
		t.Fatalf("expected __group detail element, got %s", output)
	}
	if !strings.Contains(output, "precisionlocation") {
		t.Fatalf("expected precisionlocation detail element, got %s", output)
	}
}

func TestEventValidateMissingUID(t *testing.T) {
	event := cot.Event{}
	err := event.Validate()
	if !errors.Is(err, cot.ErrUIDRequired) {
		t.Fatalf("error = %v, want %v", err, cot.ErrUIDRequired)
	}
}

func TestEventValidatePointRange(t *testing.T) {
	now := time.Now().UTC()
	event := cot.Event{
		UID:   "alpha-2",
		Type:  "a-f-G-U-C",
		How:   "m-g",
		Time:  now,
		Start: now,
		Stale: now.Add(time.Minute),
		Point: cot.Point{
			Lat: 91,
			Lon: 10,
		},
	}

	if err := event.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEventValidateQoSFormat(t *testing.T) {
	now := time.Now().UTC()
	event := cot.Event{
		Version: cot.Version2,
		UID:     "alpha-3",
		Type:    "a-f-G-U-C",
		How:     "m-g",
		Time:    now,
		Start:   now,
		Stale:   now.Add(time.Minute),
		QoS:     "bad-qos",
		Point: cot.Point{
			Lat: 10,
			Lon: 10,
		},
	}

	err := event.Validate()
	if !errors.Is(err, cot.ErrQoSInvalid) {
		t.Fatalf("error = %v, want %v", err, cot.ErrQoSInvalid)
	}
}
