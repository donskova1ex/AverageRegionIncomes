package domain

type Regions struct {
	ID         string `json:"id" db:"id"`
	RegionId   int32  `json:"RegionId" db:"RegionId"`
	RegionName string `json:"RegionName" db:"RegionName"`
}
