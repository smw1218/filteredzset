package filteredzset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/smw1218/sskiplist"
)

func TestMe(t *testing.T) {
	ss := New()
	srs := []*SaveRecord{}
	rnd := rand.New(rand.NewSource(42))
	rg := NewRandG()

	for i := 10; i >= 0; i-- {
		uuid, _ := uuid.NewRandomFromReader(rnd)

		sr := &SaveRecord{
			ID:         uuid,
			Score:      int64(i),
			FiltersRec: rg.gimme(),
		}
		srs = append(srs, sr)
		ss.Set(sr)
	}

	for filter, sl := range ss.sorts {
		fmt.Println(filter)
		sskiplist.PrintList(sl)
	}

	fmt.Println("Data")
	for k, v := range ss.data {
		fmt.Printf("%v %v\n", k, v)
	}

	fmt.Println()
	val := ss.Get(srs[4].ID, "AA")
	fmt.Printf("V %v\n", val)

	fmt.Println()
	fmt.Println("Gets")
	frecs := ss.GetAround(uuid.MustParse("ba843ee8-d63e-4c4f-be1c-ebea546d8fac"), "DD", 2, 2)
	for i, fr := range frecs {
		fmt.Printf("%d %v\n", i, fr)
	}

	fmt.Println()
	fmt.Println("Gets")
	frecs = ss.GetAround(uuid.MustParse("5b1484f2-5209-49d9-b43e-92ba09dd9d52"), "DD", 2, 2)
	for i, fr := range frecs {
		fmt.Printf("%d %v\n", i, fr)
	}
}

type RandG struct {
	groups []string
	rnd    *rand.Rand
}

func NewRandG() *RandG {
	rg := &RandG{
		groups: []string{"AA", "BB", "CC", "DD", "EE"},
		rnd:    rand.New(rand.NewSource(7)),
	}
	return rg
}

func (rg *RandG) gimme() []string {
	lengroups := len(rg.groups)
	len := rg.rnd.Intn(lengroups-1) + 1
	perm := rg.rnd.Perm(lengroups)
	rgroups := make([]string, len)
	for i, _ := range rgroups {
		rgroups[i] = rg.groups[perm[i]]
	}
	return rgroups
}
