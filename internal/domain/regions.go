package domain

type Regions struct {
	RegionId   int32  `json:"RegionId" db:"RegionId"`
	RegionName string `json:"RegionName" db:"RegionName"`
}
