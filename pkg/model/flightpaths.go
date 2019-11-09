package model

import (
	"github.com/twpayne/go-kml"
	"image/color"
	"github.com/richardmorrey/flap/pkg/flap"
	"fmt"
	"time"
	"path/filepath"
	"os"
)

type flightPaths struct {
	doc *kml.CompoundElement
	start flap.EpochTime
}

// newFlightPaths is factory function for flightPaths, struc for managing
// creation of KML representing flight paths
const maxStyles=5
func newFlightPaths(start flap.EpochTime) *flightPaths {
	fp := new(flightPaths)
	fp.doc = kml.Document(kml.Name("Flight Paths"),
        			kml.SharedStyle("0",
	                	kml.LineStyle(
			                kml.Color(color.RGBA{R: 255, G: 165, B: 0, A: 127}),
				        kml.Width(2),
					),
				),
				kml.SharedStyle("1",
	                	kml.LineStyle(
			                kml.Color(color.RGBA{R: 255, G: 255, B: 0, A: 127}),
				        kml.Width(2),
					),
				),
				kml.SharedStyle("2",
	                	kml.LineStyle(
			                kml.Color(color.RGBA{R: 255, G: 0, B: 0, A: 127}),
				        kml.Width(2),
					),
				),
				kml.SharedStyle("3",
	                	kml.LineStyle(
			                kml.Color(color.RGBA{R: 255, G: 255, B: 0, A: 127}),
				        kml.Width(2),
					),
				),
				kml.SharedStyle("4",
				kml.LineStyle(
			                kml.Color(color.RGBA{R: 255, G: 0, B: 255, A: 127}),
				        kml.Width(2),
					),
				))
	fp.start=start
	return fp
}

// addFlight adds flight path for the given flight to the KML doc
func (self *flightPaths) addFlight(fromin flap.ICAOCode,toin flap.ICAOCode,start flap.EpochTime,end flap.EpochTime, airports *flap.Airports, band bandIndex) {
	
	// Get airport locations
	from,_ := airports.GetAirport(fromin)
	to,_ := airports.GetAirport(toin)

	// Build tracker:
	gt := kml.GxTrack()
	gt.Add(kml.When(start.ToTime()))
	gt.Add(kml.When(end.ToTime()))
	gt.Add(kml.GxCoord(kml.Coordinate{Lon: from.Loc.Lon, Lat: from.Loc.Lat, Alt: 30000}))
	gt.Add(kml.GxCoord(kml.Coordinate{Lon: to.Loc.Lon, Lat: to.Loc.Lat, Alt: 30000}))
		
	// Create placemark
	style:=fmt.Sprintf("#%d",band%maxStyles)
	self.doc.Add(kml.Placemark(
		kml.Name(from.Code.ToString()+"-"+to.Code.ToString()),
		kml.StyleURL(style),
		gt))
}

// reportFlightPaths writes the current KML document out to a file
func reportFlightPaths(fp *flightPaths,end flap.EpochTime, folder string) *flightPaths {
	filename := filepath.Join(folder,fp.start.ToTime().Format(time.RFC3339)[:10] + "_" + end.ToTime().Format(time.RFC3339)[:10]+".kml")
	fh,_ := os.Create(filename)
	if fh != nil {
		kml.GxKML(fp.doc).WriteIndent(fh, "", "  ")
		fh.Close()
	}
	return newFlightPaths(end)
}

