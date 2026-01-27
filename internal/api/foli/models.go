package foli

import "encoding/xml"

// SIRIResponse represents the SIRI Stop Monitoring response
type SIRIResponse struct {
	XMLName         xml.Name        `xml:"Siri"`
	ServiceDelivery ServiceDelivery `xml:"ServiceDelivery"`
}

// ServiceDelivery contains the service delivery information
type ServiceDelivery struct {
	XMLName                xml.Name               `xml:"ServiceDelivery"`
	ResponseTimestamp      string                 `xml:"ResponseTimestamp"`
	StopMonitoringDelivery StopMonitoringDelivery `xml:"StopMonitoringDelivery"`
}

// StopMonitoringDelivery contains monitored stop visits
type StopMonitoringDelivery struct {
	XMLName             xml.Name             `xml:"StopMonitoringDelivery"`
	MonitoredStopVisits []MonitoredStopVisit `xml:"MonitoredStopVisit"`
}

// MonitoredStopVisit represents a single monitored stop visit
type MonitoredStopVisit struct {
	XMLName                 xml.Name                `xml:"MonitoredStopVisit"`
	RecordedAtTime          string                  `xml:"RecordedAtTime"`
	MonitoringRef           string                  `xml:"MonitoringRef"`
	MonitoredVehicleJourney MonitoredVehicleJourney `xml:"MonitoredVehicleJourney"`
}

// MonitoredVehicleJourney contains journey information
type MonitoredVehicleJourney struct {
	XMLName         xml.Name      `xml:"MonitoredVehicleJourney"`
	LineRef         string        `xml:"LineRef"`
	DirectionRef    string        `xml:"DirectionRef"`
	OperatorRef     string        `xml:"OperatorRef"`
	DestinationName string        `xml:"DestinationName"`
	Monitored       bool          `xml:"Monitored"`
	MonitoredCall   MonitoredCall `xml:"MonitoredCall"`
}

// MonitoredCall contains call information
type MonitoredCall struct {
	XMLName               xml.Name `xml:"MonitoredCall"`
	StopPointRef          string   `xml:"StopPointRef"`
	Order                 int      `xml:"Order"`
	StopPointName         string   `xml:"StopPointName"`
	AimedArrivalTime      string   `xml:"AimedArrivalTime"`
	ExpectedArrivalTime   string   `xml:"ExpectedArrivalTime"`
	ArrivalStatus         string   `xml:"ArrivalStatus"`
	AimedDepartureTime    string   `xml:"AimedDepartureTime"`
	ExpectedDepartureTime string   `xml:"ExpectedDepartureTime"`
	DepartureStatus       string   `xml:"DepartureStatus"`
}

// DepartureInfo represents parsed departure information
type DepartureInfo struct {
	Stop          string
	Line          string
	Destination   string
	ScheduledTime string
	ExpectedTime  string
	Status        string
	RecordedAt    string
}
