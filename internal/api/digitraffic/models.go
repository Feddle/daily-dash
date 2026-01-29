package digitraffic

// WeatherStationResponse represents the Digitraffic weather station response (GeoJSON)
type WeatherStationResponse struct {
	Type            string    `json:"type"`
	DataUpdatedTime string    `json:"dataUpdatedTime"`
	Features        []Feature `json:"features"`
}

// Feature represents a single weather station feature
type Feature struct {
	Type       string            `json:"type"`
	ID         int               `json:"id"`
	Geometry   Geometry          `json:"geometry"`
	Properties StationProperties `json:"properties"`
}

// Geometry represents geographic coordinates
type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [lon, lat, alt]
}

// StationProperties contains station properties
type StationProperties struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	CollectionStatus string `json:"collectionStatus"`
	State            string `json:"state"`
	DataUpdatedTime  string `json:"dataUpdatedTime"`
}

// RoadConditionData represents parsed road condition data
type RoadConditionData struct {
	Route       string
	Temperature float64
	Condition   string
	Location    string
	UpdatedTime string
}
