package domain

import "time"

// Weather represents weather data
type Weather struct {
	Location    string
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	Conditions  string
	Timestamp   time.Time
}

// Transit represents public transit data
type Transit struct {
	Line       string
	Stop       string
	Departures []Departure
	Timestamp  time.Time
	Warning    string
}

// Departure represents a single departure
type Departure struct {
	Stop            string
	ScheduledTime   time.Time
	ExpectedTime    time.Time
	Status          DepartureStatus
	MinutesUntil    int
	DestinationStop string
	ArrivalTime     time.Time
	DebugInfo       string
}

// DepartureStatus represents the status of a departure
type DepartureStatus int

const (
	OnTime DepartureStatus = iota
	Delayed
	Cancelled
)

func (ds DepartureStatus) String() string {
	switch ds {
	case OnTime:
		return "On Time"
	case Delayed:
		return "Delayed"
	case Cancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// RoadConditions represents road condition data
type RoadConditions struct {
	Region    string
	Segments  []RoadSegment
	Timestamp time.Time
}

// RoadSegment represents a single road segment condition
type RoadSegment struct {
	Route          string
	Condition      RoadCondition
	Temperature    float64 // Road surface temperature
	AirTemperature float64 // Air temperature
	Description    string
}

// RoadCondition represents the condition of a road
type RoadCondition int

const (
	Normal RoadCondition = iota
	Slippery
	Difficult
	Unknown
)

func (rc RoadCondition) String() string {
	switch rc {
	case Normal:
		return "Normal"
	case Slippery:
		return "Slippery"
	case Difficult:
		return "Difficult"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// DashboardData holds all dashboard data
type DashboardData struct {
	Weather        *Weather
	Transit        *Transit
	RoadConditions *RoadConditions
	Timestamp      time.Time
}
