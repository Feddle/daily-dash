package digitraffic

// WeatherStationResponse represents the Digitraffic weather station response
type WeatherStationResponse struct {
	Type       string                  `json:"type"`
	DataSource string                  `json:"dataSource"`
	Stations   []WeatherStationFeature `json:"features"`
}

// WeatherStationFeature represents a single weather station feature
type WeatherStationFeature struct {
	Type       string                   `json:"type"`
	ID         int                      `json:"id"`
	Geometry   Geometry                 `json:"geometry"`
	Properties WeatherStationProperties `json:"properties"`
}

// Geometry represents geographic coordinates
type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// WeatherStationProperties contains station properties
type WeatherStationProperties struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	CollectionStatus string        `json:"collectionStatus"`
	State            string        `json:"state"`
	DataUpdatedTime  string        `json:"dataUpdatedTime"`
	SensorValues     []SensorValue `json:"sensorValues"`
	FreeFlowSpeed1   float64       `json:"freeFlowSpeed1,omitempty"`
	FreeFlowSpeed2   float64       `json:"freeFlowSpeed2,omitempty"`
	RoadAddress      *RoadAddress  `json:"roadAddress,omitempty"`
}

// SensorValue represents a sensor reading
type SensorValue struct {
	ID                       int     `json:"id"`
	RoadStationId            int     `json:"roadStationId"`
	Name                     string  `json:"name"`
	ShortName                string  `json:"shortName"`
	TimeWindowStart          string  `json:"timeWindowStart"`
	TimeWindowEnd            string  `json:"timeWindowEnd"`
	MeasuredTime             string  `json:"measuredTime"`
	Value                    float64 `json:"value"`
	Unit                     string  `json:"unit"`
	SensorValueDescriptionEn string  `json:"sensorValueDescriptionEn,omitempty"`
	SensorValueDescriptionFi string  `json:"sensorValueDescriptionFi,omitempty"`
}

// RoadAddress represents a road address
type RoadAddress struct {
	Road            int `json:"road"`
	RoadSection     int `json:"roadSection"`
	Distance        int `json:"distance"`
	CarriagewayCode int `json:"carriagewayCode,omitempty"`
}

// RoadConditionData represents parsed road condition data
type RoadConditionData struct {
	Route       string
	Temperature float64
	Condition   string
	Location    string
	UpdatedTime string
}
