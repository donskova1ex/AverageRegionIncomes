package domain

type AverageRegionIncomes struct {
	RegionId             int32   `json:"RegionId"`
	RegionName           string  `json:"RegionName"`
	Year                 int32   `json:"Year"`
	Quarter              int32   `json:"Quarter"`
	AverageRegionIncomes float32 `json:"AverageRegionIncomes"`
}
