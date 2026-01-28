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
