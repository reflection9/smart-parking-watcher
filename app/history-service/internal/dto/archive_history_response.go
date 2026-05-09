package dto

import "time"

type ArchiveHistoryResponse struct {
	ArchivedCount int       `json:"archived_count"`
	ObjectKey     string    `json:"object_key,omitempty"`
	Cutoff        time.Time `json:"cutoff"`
}
