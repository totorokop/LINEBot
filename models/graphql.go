package models

type StationByCoordsResponse struct {
	StationByCoords StationByCoords `json:"stationByCoords"`
}

type StationByCoords struct {
	Address  string  `json:"address"`
	Lines    []Line  `json:"lines"`
	Name     string  `json:"name"`
	Distance float64 `json:"distance"`
}

type Line struct {
	Name string `json:"name"`
}
