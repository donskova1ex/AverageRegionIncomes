package domain

type RegionIncomes struct {
	ID       int32   `json:"id" db:"id"`
	RegionId int32   `json:"RegionId" db:"RegionId"`
	Year     int32   `json:"Year" db:"Year"`
	Quarter  int32   `json:"Quarter" db:"Quarter"`
	Value    float32 `json:"Value" db:"Value"`
}
