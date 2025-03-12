package domain

type RegionIncomes struct {
	ID       int32   `json:"id" db:"id"`
	RegionId int32   `json:"RegionId" db:"region_id"`
	Year     int32   `json:"Year" db:"year"`
	Quarter  int32   `json:"Quarter" db:"quarter"`
	Value    float32 `json:"Value" db:"value"`
}
