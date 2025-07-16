// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package geo_test

import (
	"math"
	"strings"
	"testing"

	"tailscale.com/types/geo"
	"tailscale.com/util/must"
)

func TestDegrees(t *testing.T) {
	for _, tt := range []struct {
		name      string
		degs      geo.Degrees
		wantStr   string
		wantText  string
		wantPad   string
		wantRads  geo.Radians
		wantTurns geo.Turns
	}{
		{
			name:      "zero",
			degs:      0.0 * geo.Degree,
			wantStr:   "+0°",
			wantText:  "+0",
			wantPad:   "+000",
			wantRads:  0.0 * geo.Radian,
			wantTurns: 0 * geo.Turn,
		},
		{
			name:      "quarter-turn",
			degs:      90.0 * geo.Degree,
			wantStr:   "+90°",
			wantText:  "+90",
			wantPad:   "+090",
			wantRads:  0.5 * math.Pi * geo.Radian,
			wantTurns: 0.25 * geo.Turn,
		},
		{
			name:      "half-turn",
			degs:      180.0 * geo.Degree,
			wantStr:   "+180°",
			wantText:  "+180",
			wantPad:   "+180",
			wantRads:  1.0 * math.Pi * geo.Radian,
			wantTurns: 0.5 * geo.Turn,
		},
		{
			name:      "full-turn",
			degs:      360.0 * geo.Degree,
			wantStr:   "+360°",
			wantText:  "+360",
			wantPad:   "+360",
			wantRads:  2 * math.Pi * geo.Radian,
			wantTurns: 1.0 * geo.Turn,
		},
		{
			name:      "negative-zero",
			degs:      must.Get(geo.ParseDegrees("-0.0")),
			wantStr:   "-0°",
			wantText:  "-0",
			wantPad:   "-000",
			wantRads:  0 * geo.Radian * -1,
			wantTurns: 0 * geo.Turn * -1,
		},
		{
			name:      "small-degree",
			degs:      -1.2003 * geo.Degree,
			wantStr:   "-1.2003°",
			wantText:  "-1.2003",
			wantPad:   "-001.2003",
			wantRads:  -0.020949187011687936 * geo.Radian,
			wantTurns: -0.0033341666666666667 * geo.Turn,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.degs.String(); got != tt.wantStr {
				t.Errorf("got %q, want %q", got, tt.wantStr)
			}

			b, err := tt.degs.AppendText(nil)
			if err != nil {
				t.Fatalf("err %q, want nil", err.Error())
			}
			if string(b) != tt.wantText {
				t.Errorf("got %q, want %q", b, tt.wantText)
			}

			b = tt.degs.AppendZeroPaddedText(nil, 3)
			if string(b) != tt.wantPad {
				t.Errorf("got %q, want %q", b, tt.wantPad)
			}

			r := tt.degs.Radians()
			if r != tt.wantRads {
				t.Errorf("got %v, want %v", r, tt.wantRads)
			}

			// Roundtrip
			if d := r.Degrees(); d != tt.degs {
				t.Errorf("got %v, want %v", d, tt.degs)
			}

			ts := r.Turns()
			if ts != tt.wantTurns {
				t.Errorf("got %v, want %v", ts, tt.wantTurns)
			}

		})
	}
}

func TestDistanceOnEarth(t *testing.T) {
	for _, tt := range []struct {
		name    string
		here    geo.Point
		there   geo.Point
		want    geo.Distance
		wantErr string
	}{
		{
			name:    "no-points",
			here:    geo.Point{},
			there:   geo.Point{},
			wantErr: "not a valid point",
		},
		{
			name:    "not-here",
			here:    geo.Point{},
			there:   geo.MakePoint(0, 0),
			wantErr: "not a valid point",
		},
		{
			name:    "not-there",
			here:    geo.MakePoint(0, 0),
			there:   geo.Point{},
			wantErr: "not a valid point",
		},
		{
			name:  "null-island",
			here:  geo.MakePoint(0, 0),
			there: geo.MakePoint(0, 0),
			want:  0 * geo.Meter,
		},
		{
			name:  "equator-to-south-pole",
			here:  geo.MakePoint(0, 0),
			there: geo.MakePoint(-90, 0),
			want:  geo.EarthMeanCircumference / 4,
		},
		{
			name:  "north-pole-to-south-pole",
			here:  geo.MakePoint(+90, 0),
			there: geo.MakePoint(-90, 0),
			want:  geo.EarthMeanCircumference / 2,
		},
		{
			name:  "meridian-to-antimeridian",
			here:  geo.MakePoint(0, 0),
			there: geo.MakePoint(0, -180),
			want:  geo.EarthMeanCircumference / 2,
		},
		{
			name:  "positive-to-negative-antimeridian",
			here:  geo.MakePoint(0, 180),
			there: geo.MakePoint(0, -180),
			want:  0 * geo.Meter,
		},
		{
			name:  "toronto-to-montreal",
			here:  geo.MakePoint(+43.70011, -79.41630),
			there: geo.MakePoint(+45.50884, -73.58781),
			want:  503_200 * geo.Meter,
		},
		{
			name:  "montreal-to-san-francisco",
			here:  geo.MakePoint(+45.50884, -73.58781),
			there: geo.MakePoint(+37.77493, -122.41942),
			want:  4_082_600 * geo.Meter,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.here.DistanceTo(tt.there)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("err %q, want nil", err)
			}
			if tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err %q, want %q", err, tt.wantErr)
			}

			approx := func(x, y geo.Distance) bool {
				return math.Abs(float64(x)-float64(y)) <= 10
			}
			if !approx(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
