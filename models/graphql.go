package models

type NearbyStationsResponse struct {
	NearbyStations []NearbyStations `json:"nearbyStations"`
}

type NearbyStations struct {
	Address  string  `json:"address"`
	Lines    []Line  `json:"lines"`
	Name     string  `json:"name"`
	Distance float64 `json:"distance"`
}

type Line struct {
	Name string `json:"name"`
}
