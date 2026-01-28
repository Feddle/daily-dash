package fmi

import (
	"testing"
)

func TestParseWeatherResponse(t *testing.T) {
	// Sample XML based on fmi::observations::weather::timevaluepair response structure
	xmlData := []byte(`
<wfs:FeatureCollection xmlns:wfs="http://www.opengis.net/wfs/2.0" xmlns:om="http://www.opengis.net/om/2.0" xmlns:gml="http://www.opengis.net/gml/3.2">
	<wfs:member>
		<om:PointTimeSeriesObservation gml:id="obs-1">
			<om:observedProperty xlink:href="http://xml.fmi.fi/schema/wfs/2.0/Query/StoredQuery/fmi::observations::weather::timevaluepair#t2m"/>
			<om:result>
				<wml2:MeasurementTimeseries gml:id="mts-1">
					<wml2:point>
						<wml2:MeasurementTVP>
							<wml2:time>2024-01-28T10:00:00Z</wml2:time>
							<wml2:value>-5.2</wml2:value>
						</wml2:MeasurementTVP>
					</wml2:point>
				</wml2:MeasurementTimeseries>
			</om:result>
		</om:PointTimeSeriesObservation>
	</wfs:member>
	<wfs:member>
		<om:PointTimeSeriesObservation gml:id="obs-2">
			<om:observedProperty xlink:href="http://xml.fmi.fi/schema/wfs/2.0/Query/StoredQuery/fmi::observations::weather::timevaluepair#rh"/>
			<om:result>
				<wml2:MeasurementTimeseries gml:id="mts-2">
					<wml2:point>
						<wml2:MeasurementTVP>
							<wml2:time>2024-01-28T10:00:00Z</wml2:time>
							<wml2:value>88.0</wml2:value>
						</wml2:MeasurementTVP>
					</wml2:point>
				</wml2:MeasurementTimeseries>
			</om:result>
		</om:PointTimeSeriesObservation>
	</wfs:member>
	<wfs:member>
		<om:PointTimeSeriesObservation gml:id="obs-3">
			<om:observedProperty xlink:href="http://xml.fmi.fi/schema/wfs/2.0/Query/StoredQuery/fmi::observations::weather::timevaluepair#ws_10min"/>
			<om:result>
				<wml2:MeasurementTimeseries gml:id="mts-3">
					<wml2:point>
						<wml2:MeasurementTVP>
							<wml2:time>2024-01-28T10:00:00Z</wml2:time>
							<wml2:value>3.5</wml2:value>
						</wml2:MeasurementTVP>
					</wml2:point>
				</wml2:MeasurementTimeseries>
			</om:result>
		</om:PointTimeSeriesObservation>
	</wfs:member>
</wfs:FeatureCollection>`)

	obs, timestamp, err := ParseWeatherResponse(xmlData)
	if err != nil {
		t.Fatalf("ParseWeatherResponse failed: %v", err)
	}

	if timestamp != "2024-01-28T10:00:00Z" {
		t.Errorf("Expected timestamp 2024-01-28T10:00:00Z, got %s", timestamp)
	}

	data := ExtractWeatherData(obs, timestamp)

	if data.Temperature != -5.2 {
		t.Errorf("Expected temperature -5.2, got %f", data.Temperature)
	}
	if data.Humidity != 88.0 {
		t.Errorf("Expected humidity 88.0, got %f", data.Humidity)
	}
	if data.WindSpeed != 3.5 {
		t.Errorf("Expected wind speed 3.5, got %f", data.WindSpeed)
	}
}

func TestParseWeatherResponse_RealAPIFormat(t *testing.T) {
	// Sample XML from real API response (using query params in href)
	xmlData := []byte(`
<wfs:FeatureCollection xmlns:wfs="http://www.opengis.net/wfs/2.0" xmlns:om="http://www.opengis.net/om/2.0" xmlns:gml="http://www.opengis.net/gml/3.2" xmlns:xlink="http://www.w3.org/1999/xlink">
	<wfs:member>
		<om:PointTimeSeriesObservation gml:id="obs-1">
			<om:observedProperty xlink:href="http://opendata.fmi.fi/meta?observableProperty=observation&amp;param=t2m&amp;language=eng"/>
			<om:result>
				<wml2:MeasurementTimeseries gml:id="mts-1" xmlns:wml2="http://www.opengis.net/waterml/2.0">
					<wml2:point>
						<wml2:MeasurementTVP>
							<wml2:time>2024-01-28T10:00:00Z</wml2:time>
							<wml2:value>-5.2</wml2:value>
						</wml2:MeasurementTVP>
					</wml2:point>
				</wml2:MeasurementTimeseries>
			</om:result>
		</om:PointTimeSeriesObservation>
	</wfs:member>
	<wfs:member>
		<om:PointTimeSeriesObservation gml:id="obs-2">
			<om:observedProperty xlink:href="http://opendata.fmi.fi/meta?observableProperty=observation&amp;param=rh&amp;language=eng"/>
			<om:result>
				<wml2:MeasurementTimeseries gml:id="mts-2" xmlns:wml2="http://www.opengis.net/waterml/2.0">
					<wml2:point>
						<wml2:MeasurementTVP>
							<wml2:time>2024-01-28T10:00:00Z</wml2:time>
							<wml2:value>88.0</wml2:value>
						</wml2:MeasurementTVP>
					</wml2:point>
				</wml2:MeasurementTimeseries>
			</om:result>
		</om:PointTimeSeriesObservation>
	</wfs:member>
</wfs:FeatureCollection>`)

	obs, _, err := ParseWeatherResponse(xmlData)
	if err != nil {
		t.Fatalf("ParseWeatherResponse failed: %v", err)
	}

	if t2m, ok := obs["t2m"]; !ok || t2m != -5.2 {
		t.Errorf("Expected t2m -5.2, got %v", obs["t2m"])
	}
	if rh, ok := obs["rh"]; !ok || rh != 88.0 {
		t.Errorf("Expected rh 88.0, got %v", obs["rh"])
	}
}
