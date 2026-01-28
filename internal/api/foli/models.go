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
