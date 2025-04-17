package domain

type AverageRegionIncomes struct {
	RegionId             int32   `db:"region_id" json:"RegionId"`
	RegionName           string  `db:"region_name" json:"RegionName"`
	Year                 int32   `db:"year" json:"Year"`
	Quarter              int32   `db:"quarter" json:"Quarter"`
	AverageRegionIncomes float32 `db:"average_region_incomes" json:"AverageRegionIncomes"`
}
