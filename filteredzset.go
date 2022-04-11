package filteredzset

import (
	"fmt"
	"io"
	"sort"

	"github.com/smw1218/sskiplist"
)

type SortedSet[T FilteredOrderable[T]] struct {
	data map[interface{}][]*element[T]
	// keyed by the Filter values
	sorts map[string]*sskiplist.SL[T]
}

type FilteredOrderable[T any] interface {
	sskiplist.Orderable[T]
	Filters() []string
	Key() interface{}
}

type FilteredRecord[T FilteredOrderable[T]] struct {
	Key       interface{}
	Value     T
	Filter    string
	Index     int
	Total     int
	Requested bool
}

func (fr *FilteredRecord[T]) String() string {
	return fmt.Sprintf("%s %d/%d %5v|%v", fr.Filter, fr.Index, fr.Total, fr.Requested, fr.Value)
}

type element[T sskiplist.Orderable[T]] struct {
	Filter   string
	Skiplist *sskiplist.SL[T]
	Element  *sskiplist.Element[T]
	Index    int
}

func (e *element[T]) String() string {
	return fmt.Sprintf("%s:%d", e.Filter, e.Index)
}

func New[T FilteredOrderable[T]]() *SortedSet[T] {
	return &SortedSet[T]{
		data:  make(map[interface{}][]*element[T]),
		sorts: make(map[string]*sskiplist.SL[T]),
	}
}

func (ss *SortedSet[T]) Set(sr T) {
	filters := sr.Filters()
	retchans := make([]chan *indexAndElement[T], len(filters))
	newElements := make([]*element[T], len(filters))
	oldRecords := ss.data[sr.Key()]
	for i, filter := range filters {
		sl, ok := ss.sorts[filter]
		if !ok {
			sl = sskiplist.New[T]()
			ss.sorts[filter] = sl
		}
		oldElement := getFilterRecord(filter, oldRecords)
		retchans[i] = make(chan *indexAndElement[T])
		go ss.asyncUpdateSkiplist(oldElement, sl, sr, retchans[i])
	}
	// TODO remove where oldRecord.Filter exist but missing in the new sr.Filters

	for i := range filters {
		ret := <-retchans[i]
		newElements[i] = &element[T]{
			Filter:   filters[i],
			Skiplist: ss.sorts[filters[i]],
			Element:  ret.Element,
			Index:    ret.Index,
		}
	}
	ss.data[sr.Key()] = newElements
}

func getFilterRecord[T sskiplist.Orderable[T]](filter string, records []*element[T]) *element[T] {
	for _, e := range records {
		if e.Filter == filter {
			return e
		}
	}
	return nil
}

type indexAndElement[T FilteredOrderable[T]] struct {
	Index   int
	Element *sskiplist.Element[T]
}

func (ss *SortedSet[T]) Summary(w io.Writer) {
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

func (ss *SortedSet[T]) Size(filter string) int {
	sl, ok := ss.sorts[filter]
	if !ok {
		return 0
	}
	return sl.Size()
}

func (ss *SortedSet[T]) asyncUpdateSkiplist(oldRecord *element[T], sl *sskiplist.SL[T], newRecord T, echan chan *indexAndElement[T]) {
	if oldRecord != nil {
		sl.Remove(oldRecord.Element.Value)
	}
	idx, e := sl.Set(newRecord)
	echan <- &indexAndElement[T]{idx, e}
}
func (ss *SortedSet[T]) get(key interface{}, filter string) (*FilteredRecord[T], *element[T]) {
	elements, ok := ss.data[key]
	if !ok {
		return nil, nil
	}
	var filterElement *element[T]
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

	sr := filterElement.Element.Value
	fr := filteredRecord(idx, filterElement.Skiplist.Size(), sr, filter)
	fr.Requested = true
	return fr, filterElement
}

func (ss *SortedSet[T]) Get(key interface{}, filter string) *FilteredRecord[T] {
	fr, _ := ss.get(key, filter)
	return fr
}

func (ss *SortedSet[T]) GetAround(key interface{}, filter string, before, after int) []*FilteredRecord[T] {
	fr, filterElement := ss.get(key, filter)
	if fr == nil {
		return []*FilteredRecord[T]{}
	}

	filterRecords := make([]*FilteredRecord[T], before+after+1)
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
		sr := ele.Value
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
		sr := ele.Value
		filterRecords[last] = filteredRecord(idx+i+1, filterElement.Skiplist.Size(), sr, filter)
		last++
	}

	return filterRecords[first:last]
	//return filterRecords
}

func filteredRecord[T FilteredOrderable[T]](idx, total int, sr T, filter string) *FilteredRecord[T] {
	return &FilteredRecord[T]{
		Key:       sr.Key(),
		Value:     sr,
		Filter:    filter,
		Index:     idx,
		Total:     total,
		Requested: false,
	}
}
