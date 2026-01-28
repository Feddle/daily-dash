package foli

// SIRIJSONResponse represents the Föli SIRI JSON response
type SIRIJSONResponse struct {
	Sys    string             `json:"sys"`
	Status string             `json:"status"`
	Result []VehicleDeparture `json:"result"`
}

// VehicleDeparture represents a single vehicle departure/status in JSON
type VehicleDeparture struct {
	LineRef               string      `json:"lineref"`
	DestinationDisplay    string      `json:"destinationdisplay"`
	ExpectedDepartureTime int64       `json:"expecteddeparturetime"` // Unix timestamp
	AimedDepartureTime    int64       `json:"aimeddeparturetime"`    // Unix timestamp
	RecordedAtTime        int64       `json:"recordedattime"`
	Delay                 interface{} `json:"delay"`
}

// DepartureInfo represents parsed departure information for consumption by the coordinator
type DepartureInfo struct {
	Stop          string
	Line          string
	Destination   string
	ScheduledTime string
	ExpectedTime  string
	Status        string
	RecordedAt    string
}

// GTFSStop represents a bus stop from the Föli GTFS API
type GTFSStop struct {
	ID   string `json:"stop_id"`
	Name string `json:"stop_name"`
	Code string `json:"stop_code"`
	Desc string `json:"stop_desc"`
}

// FilterValue implements the list.Item interface
func (s GTFSStop) FilterValue() string {
	return s.Name
}

// Title implements the list.Item interface
func (s GTFSStop) Title() string {
	return s.Name
}

// Description implements the list.Item interface
func (s GTFSStop) Description() string {
	return s.Desc
}
