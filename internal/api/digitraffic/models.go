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

// -- Forecast API Models --

// ForecastResponse represents the road condition forecast response
type ForecastResponse struct {
	DataUpdatedTime  string            `json:"dataUpdatedTime"`
	ForecastSections []ForecastSection `json:"forecastSections"`
}

// ForecastSection represents a section of road with weather forecasts
type ForecastSection struct {
	ID        string     `json:"id"`
	Forecasts []Forecast `json:"forecasts"`
}

// Forecast represents a single forecast entry
type Forecast struct {
	Time                    string          `json:"time"`
	Type                    string          `json:"type"` // OBSERVATION or FORECAST
	OverallRoadCondition    string          `json:"overallRoadCondition"`
	RoadTemperature         float64         `json:"roadTemperature"`
	Temperature             float64         `json:"temperature"`
	WindSpeed               float64         `json:"windSpeed"`
	ForecastConditionReason ConditionReason `json:"forecastConditionReason"`
	Reliability             string          `json:"reliability"`
}

// ConditionReason contains details about the condition
type ConditionReason struct {
	RoadCondition string `json:"roadCondition"` // SNOW, ICE, etc.
}

// ForecastMetadataResponse represents the metadata for forecast sections (names/descriptions)
type ForecastMetadataResponse struct {
	Features []MetadataFeature `json:"features"`
}

type MetadataFeature struct {
	Properties MetadataProperties `json:"properties"`
}

type MetadataProperties struct {
	ID          string `json:"id"`
	Description string `json:"description"` // e.g. "Vt 1: Helsinki - Turku"
	RoadNumber  int    `json:"roadNumber"`
}

// RoadConditionData represents parsed road condition data
type RoadConditionData struct {
	Route          string
	Temperature    float64 // Road surface temperature
	AirTemperature float64 // Air temperature
	Condition      string
	Location       string
	UpdatedTime    string
}
