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

func (sr *SaveRecord) Less(other interface{}) bool {
	osr := other.(*SaveRecord)
	if sr.Score == osr.Score {
		return sr.ID[0] < osr.ID[0]
	}
	return sr.Score > osr.Score
}

func (sr *SaveRecord) Equal(other interface{}) bool {
	osr := other.(*SaveRecord)
	return sr.Score == osr.Score && sr.ID[0] == osr.ID[0]
}

func (sr *SaveRecord) Filters() []string {
	return sr.FiltersRec
}

func (sr *SaveRecord) Key() interface{} {
	return sr.ID
}
