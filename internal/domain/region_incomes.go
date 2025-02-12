package domain

type RegionIncomes struct {
	RegionId             int32   `json:"RegionId" db:"RegionId"`
	Year                 int32   `json:"Year" db:"Year"`
	Quarter              int32   `json:"Quarter" db:"Quarter"`
	AverageRegionIncomes float32 `json:"AverageRegionIncomes" db:"AverageRegionIncomes"`
}
