package model

type SpotStatus string

const (
	SpotStatusFree         SpotStatus = "FREE"
	SpotStatusOccupied     SpotStatus = "OCCUPIED"
	SpotStatusBlocked      SpotStatus = "BLOCKED"
	SpotStatusOutOfService SpotStatus = "OUT_OF_SERVICE"
)

func IsValidSpotStatus(status string) bool {
	switch SpotStatus(status) {
	case SpotStatusFree, SpotStatusOccupied, SpotStatusBlocked, SpotStatusOutOfService:
		return true
	default:
		return false
	}
}
