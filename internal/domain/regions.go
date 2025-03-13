package domain

type Regions struct {
	ID         string `json:"id" db:"id"`
	RegionId   int32  `json:"RegionId" db:"region_id"`
	RegionName string `json:"RegionName" db:"region_name"`
}
