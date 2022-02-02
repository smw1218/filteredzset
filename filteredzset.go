package filteredzset

import (
	"fmt"
	"io"
	"sort"

	"github.com/smw1218/sskiplist"
)

type SortedSet struct {
	data map[interface{}][]*element
	// keyed by the Filter values
	sorts map[string]*sskiplist.SL
}

type FilteredOrderable interface {
	sskiplist.Orderable
	Filters() []string
	Key() interface{}
}

type FilteredRecord struct {
	Key       interface{}
	Value     interface{}
	Filter    string
	Index     int
	Total     int
	Requested bool
}

func (fr *FilteredRecord) String() string {
	return fmt.Sprintf("%s %d/%d %5v|%v", fr.Filter, fr.Index, fr.Total, fr.Requested, fr.Value)
}

type element struct {
	Filter   string
	Skiplist *sskiplist.SL
	Element  *sskiplist.Element
	Index    int
}

func (e *element) String() string {
	return fmt.Sprintf("%s:%d", e.Filter, e.Index)
}

func New() *SortedSet {
	return &SortedSet{
		data:  make(map[interface{}][]*element),
		sorts: make(map[string]*sskiplist.SL),
	}
}

func (ss *SortedSet) Set(sr FilteredOrderable) {
	filters := sr.Filters()
	retchans := make([]chan *indexAndElement, len(filters))
	newElements := make([]*element, len(filters))
	oldRecords := ss.data[sr.Key()]
	for i, filter := range filters {
		sl, ok := ss.sorts[filter]
		if !ok {
			sl = sskiplist.New()
			ss.sorts[filter] = sl
		}
		oldElement := getFilterRecord(filter, oldRecords)
		retchans[i] = make(chan *indexAndElement)
		go ss.asyncUpdateSkiplist(oldElement, sl, sr, retchans[i])
	}
	// TODO remove where oldRecord.Filter exist but missing in the new sr.Filters

	for i := range filters {
		ret := <-retchans[i]
		newElements[i] = &element{
			Filter:   filters[i],
			Skiplist: ss.sorts[filters[i]],
			Element:  ret.Element,
			Index:    ret.Index,
		}
	}
	ss.data[sr.Key()] = newElements
}

func getFilterRecord(filter string, records []*element) *element {
	for _, e := range records {
		if e.Filter == filter {
			return e
		}
	}
	return nil
}

type indexAndElement struct {
	Index   int
	Element *sskiplist.Element
}

func (ss *SortedSet) Summary(w io.Writer) {
	fmt.Fprintf(w, "All Records: %d\n", len(ss.data))
	fmt.Fprintf(w, "Filters: %d\n", len(ss.sorts))
	keys := make([]string, len(ss.sorts))
	i := 0
	for k, _ := range ss.sorts {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := ss.sorts[k]
		fmt.Fprintf(w, "%s: %d\n", k, v.Size())
	}
}

func (ss *SortedSet) Size(filter string) int {
	sl, ok := ss.sorts[filter]
	if !ok {
		return 0
	}
	return sl.Size()
}

func (ss *SortedSet) asyncUpdateSkiplist(oldRecord *element, sl *sskiplist.SL, newRecord FilteredOrderable, echan chan *indexAndElement) {
	if oldRecord != nil {
		sl.Remove(oldRecord.Element.Value)
	}
	idx, e := sl.Set(newRecord)
	echan <- &indexAndElement{idx, e}
}
func (ss *SortedSet) get(key interface{}, filter string) (*FilteredRecord, *element) {
	elements, ok := ss.data[key]
	if !ok {
		return nil, nil
	}
	var filterElement *element
	for _, e := range elements {
		if e.Filter == filter {
			filterElement = e
			break
		}
	}
	if filterElement == nil {
		return nil, nil
	}
	idx, _ := filterElement.Skiplist.Get(filterElement.Element.Value)

	sr := filterElement.Element.Value.(FilteredOrderable)
	fr := filteredRecord(idx, filterElement.Skiplist.Size(), sr, filter)
	fr.Requested = true
	return fr, filterElement
}

func (ss *SortedSet) Get(key interface{}, filter string) *FilteredRecord {
	fr, _ := ss.get(key, filter)
	return fr
}

func (ss *SortedSet) GetAround(key interface{}, filter string, before, after int) []*FilteredRecord {
	fr, filterElement := ss.get(key, filter)
	if fr == nil {
		return []*FilteredRecord{}
	}

	filterRecords := make([]*FilteredRecord, before+after+1)
	filterRecords[before] = fr
	idx := fr.Index

	// grab the before elements
	first := before
	ele := filterElement.Element
	for i := 0; i < before; i++ {
		ele = ele.Prev()
		if ele == nil {
			break
		}
		sr := ele.Value.(FilteredOrderable)
		filterRecords[first-1] = filteredRecord(idx-i-1, filterElement.Skiplist.Size(), sr, filter)
		first--
	}

	// grab the after elements
	last := before + 1
	ele = filterElement.Element
	for i := 0; i < after; i++ {
		ele = ele.Next()
		if ele == nil {
			break
		}
		sr := ele.Value.(FilteredOrderable)
		filterRecords[last] = filteredRecord(idx+i+1, filterElement.Skiplist.Size(), sr, filter)
		last++
	}

	return filterRecords[first:last]
	//return filterRecords
}

func filteredRecord(idx, total int, sr FilteredOrderable, filter string) *FilteredRecord {
	return &FilteredRecord{
		Key:       sr.Key(),
		Value:     sr,
		Filter:    filter,
		Index:     idx,
		Total:     total,
		Requested: false,
	}
}
