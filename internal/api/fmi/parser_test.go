package fmi

import (
	"encoding/xml"
	"testing"
)

func TestParseStationsResponse(t *testing.T) {
	// Sample XML from FMI API
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<wfs:FeatureCollection xmlns:wfs="http://www.opengis.net/wfs/2.0" xmlns:gml="http://www.opengis.net/gml/3.2" xmlns:ins_base="http://inspire.ec.europa.eu/schemas/base/3.3" xmlns:ef="http://inspire.ec.europa.eu/schemas/ef/4.0">
  <wfs:member>
    <ef:EnvironmentalMonitoringFacility>
      <ef:inspireId>
        <ins_base:Identifier>
          <ins_base:localId>100971</ins_base:localId>
        </ins_base:Identifier>
      </ef:inspireId>
      <ef:name>Helsinki Kaisaniemi</ef:name>
      <ef:representativePoint>
        <gml:Point>
          <gml:pos>60.17523 24.94459</gml:pos>
        </gml:Point>
      </ef:representativePoint>
      <ef:operationalActivityPeriod>
        <ef:OperationalActivityPeriod>
          <ef:activityTime>
            <gml:TimePeriod>
              <gml:beginPosition>1844-11-01T00:00:00Z</gml:beginPosition>
              <gml:endPosition indeterminatePosition="now"/>
            </gml:TimePeriod>
          </ef:activityTime>
        </ef:OperationalActivityPeriod>
      </ef:operationalActivityPeriod>
    </ef:EnvironmentalMonitoringFacility>
  </wfs:member>
</wfs:FeatureCollection>`

	stations, err := ParseStationsResponse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Failed to parse stations: %v", err)
	}

	if len(stations) != 1 {
		t.Fatalf("Expected 1 station, got %d", len(stations))
	}

	station := stations[0]
	if station.Name != "Helsinki Kaisaniemi" {
		t.Errorf("Expected name 'Helsinki Kaisaniemi', got '%s'", station.Name)
	}
	if station.FMISID != "100971" {
		t.Errorf("Expected FMISID '100971', got '%s'", station.FMISID)
	}
	if !station.Active {
		t.Error("Expected station to be active")
	}
}

func TestXMLParsing(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<wfs:FeatureCollection xmlns:wfs="http://www.opengis.net/wfs/2.0" xmlns:gml="http://www.opengis.net/gml/3.2" xmlns:ins_base="http://inspire.ec.europa.eu/schemas/base/3.3" xmlns:ef="http://inspire.ec.europa.eu/schemas/ef/4.0">
  <wfs:member>
    <ef:EnvironmentalMonitoringFacility>
      <ef:inspireId>
        <ins_base:Identifier>
          <ins_base:localId>100971</ins_base:localId>
        </ins_base:Identifier>
      </ef:inspireId>
      <ef:name>Helsinki Kaisaniemi</ef:name>
      <ef:representativePoint>
        <gml:Point>
          <gml:pos>60.17523 24.94459</gml:pos>
        </gml:Point>
      </ef:representativePoint>
      <ef:operationalActivityPeriod>
        <ef:OperationalActivityPeriod>
          <ef:activityTime>
            <gml:TimePeriod>
              <gml:beginPosition>1844-11-01T00:00:00Z</gml:beginPosition>
              <gml:endPosition indeterminatePosition="now"/>
            </gml:TimePeriod>
          </ef:activityTime>
        </ef:OperationalActivityPeriod>
      </ef:operationalActivityPeriod>
    </ef:EnvironmentalMonitoringFacility>
  </wfs:member>
</wfs:FeatureCollection>`

	var response StationsWFSResponse
	err := xml.Unmarshal([]byte(xmlData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	if len(response.Members) != 1 {
		t.Fatalf("Expected 1 member, got %d", len(response.Members))
	}

	facility := response.Members[0].Facility
	t.Logf("Name: %s", facility.Name)
	t.Logf("FMISID: %s", facility.InspireID.Identifier.LocalID)
	t.Logf("Pos: %s", facility.RepresentativePoint.Point.Pos)
	t.Logf("End: '%s'", facility.Period.Period.ActivityTime.TimePeriod.End)
}
