# Sorted Set with multiple Groupings

This is a prototype and not ready for production use.

This is a pure go sorted set implementation. Like a [Redis zset](https://redis.io/topics/data-types#sorted-sets). This includes an additional capability to have multiple "Filters" on for each item in the set which each correspond to a different index. For instance, if you have zset with cars sorted by horsepower, then a filter on all cars which would create a list with all the cars which is similar to any regular zset. This supports adding additional filters as well which will each have their own ordering. For example, each car could return something like `[]string{"All","Tesla"}`. This would create indexing for each unique manufacturer which would allow checking the ordering for each specific manufacturer.

Items in the filtered zset need to implement the `FilteredOrderable` interface:

    type FilteredOrderable interface {
        sskiplist.Orderable
        Filters() []string
        Key() interface{}
    }

where `sskiplist.Orderable` is:

    type Orderable interface {
        Less(other interface{}) bool
        Equal(other interface{}) bool
    }

The value implementing the FilteredOrderable is stored in the SortedSet.