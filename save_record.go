package filteredzset

import (
	"fmt"

	"github.com/google/uuid"
)

// SaveRecord is an example struct implementing
// FilteredOrderable
type SaveRecord struct {
	ID         uuid.UUID
	Score      int64
	FiltersRec []string
}

func (sr *SaveRecord) String() string {
	return fmt.Sprintf("%v Score %2d in %v", sr.ID, sr.Score, sr.FiltersRec)
}

func (sr *SaveRecord) Less(other *SaveRecord) bool {

	if sr.Score == other.Score {
		return sr.ID[0] < other.ID[0]
	}
	return sr.Score > other.Score
}

func (sr *SaveRecord) Equal(other *SaveRecord) bool {
	return sr.Score == other.Score && sr.ID[0] == other.ID[0]
}

func (sr *SaveRecord) Filters() []string {
	return sr.FiltersRec
}

func (sr *SaveRecord) Key() interface{} {
	return sr.ID
}
