package fmi

import "encoding/xml"

// WFSResponse represents the FMI WFS XML response structure
type WFSResponse struct {
	XMLName xml.Name `xml:"FeatureCollection"`
	Members []Member `xml:"member"`
}

// Member represents a single observation member
type Member struct {
	XMLName                    xml.Name                   `xml:"member"`
	PointTimeSeriesObservation PointTimeSeriesObservation `xml:"PointTimeSeriesObservation"`
}

// PointTimeSeriesObservation represents a time series observation
type PointTimeSeriesObservation struct {
	XMLName          xml.Name         `xml:"PointTimeSeriesObservation"`
	Result           Result           `xml:"result"`
	ObservedProperty ObservedProperty `xml:"observedProperty"`
}

// ObservedProperty contains the parameter definition
type ObservedProperty struct {
	XMLName xml.Name `xml:"observedProperty"`
	Href    string   `xml:"href,attr"`
	Title   string   `xml:"title,attr"`
}

// Result contains the measurement time series
type Result struct {
	XMLName               xml.Name              `xml:"result"`
	MeasurementTimeseries MeasurementTimeseries `xml:"MeasurementTimeseries"`
}

// MeasurementTimeseries contains the points
type MeasurementTimeseries struct {
	XMLName xml.Name              `xml:"MeasurementTimeseries"`
	Points  []MeasurementTVPPoint `xml:"point"`
}

// MeasurementTVPPoint represents a time-value pair point
type MeasurementTVPPoint struct {
	XMLName xml.Name       `xml:"point"`
	TVP     MeasurementTVP `xml:"MeasurementTVP"`
}

// MeasurementTVP contains time and value
type MeasurementTVP struct {
	XMLName xml.Name `xml:"MeasurementTVP"`
	Time    string   `xml:"time"`
	Value   float64  `xml:"value"`
}

// ObservationData represents parsed weather observation data
type ObservationData struct {
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	Time        string
}

// StationsWFSResponse represents the FMI stations response
type StationsWFSResponse struct {
	XMLName xml.Name        `xml:"FeatureCollection"`
	Members []StationMember `xml:"member"`
}

// StationMember represents a single station member
type StationMember struct {
	XMLName  xml.Name                        `xml:"member"`
	Facility EnvironmentalMonitoringFacility `xml:"EnvironmentalMonitoringFacility"`
}

// EnvironmentalMonitoringFacility represents a monitoring facility
type EnvironmentalMonitoringFacility struct {
	InspireID           InspireID                        `xml:"inspireId"`
	Name                string                           `xml:"name"`
	RepresentativePoint RepresentativePoint              `xml:"representativePoint"`
	Period              OperationalActivityPeriodWrapper `xml:"operationalActivityPeriod"`
}

// InspireID represents the inspire identifier
type InspireID struct {
	Identifier Identifier `xml:"Identifier"`
}

// Identifier contains the local ID (FMISID)
type Identifier struct {
	LocalID string `xml:"localId"` // This is the FMISID
}

// RepresentativePoint represents the station location
type RepresentativePoint struct {
	Point Point `xml:"Point"`
}

// Point represents a geographic point
type Point struct {
	Pos string `xml:"pos"` // "lat lon" format
}

// OperationalActivityPeriodWrapper wraps the operational period
type OperationalActivityPeriodWrapper struct {
	Period OperationalActivityPeriod `xml:"OperationalActivityPeriod"`
}

// OperationalActivityPeriod represents the operational period
type OperationalActivityPeriod struct {
	ActivityTime ActivityTime `xml:"activityTime"`
}

// ActivityTime contains the time period
type ActivityTime struct {
	TimePeriod TimePeriod `xml:"TimePeriod"`
}

// TimePeriod contains begin and end positions
type TimePeriod struct {
	Begin string `xml:"beginPosition"`
	End   string `xml:"endPosition"` // Empty or has indeterminatePosition="now" for active stations
}

// WeatherStation represents a weather observation station
type WeatherStation struct {
	Name      string
	FMISID    string
	Latitude  float64 // For future enhancements
	Longitude float64 // For future enhancements
	Active    bool    // Filter out inactive stations
}
