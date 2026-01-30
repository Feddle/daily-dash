package fmi

// FallbackStations is used only when API fetch fails
// These are major weather observation stations in Finland
var FallbackStations = []WeatherStation{
	{Name: "Helsinki Kaisaniemi", FMISID: "100971", Active: true},
	{Name: "Tampere Pirkkala", FMISID: "101118", Active: true},
	{Name: "Turku Artukainen", FMISID: "100949", Active: true},
	{Name: "Oulu Oulunsalo", FMISID: "101786", Active: true},
	{Name: "Rovaniemi Apukka", FMISID: "101920", Active: true},
}
