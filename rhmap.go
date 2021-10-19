package synx

import "errors"

var (
	ErrMapFull  = errors.New("map is full")
	ErrNotFound = errors.New("key not found")
)

const defaultCapacity = 10

// Map is a hashmap that uses the robinhood algorithm.
//
// A general purpose hash table, using the Robin Hood hashing algorithm.
//
// Conceptually, it is a hash table using linear probing on lookup with
// a particular displacement strategy on inserts.  The central idea of
// the Robin Hood hashing algorithm is to reduce the variance of the
// probe sequence length (PSL).
//
// Reference:
//
// Pedro Celis, 1986, Robin Hood Hashing, University of Waterloo
// https://cs.uwaterloo.ca/research/tr/1986/CS-86-14.pdf
//
type Map struct {
	// items are the slots of the hashmap for items.
	items []Item

	// Number of keys in the Map.
	count int

	// When any item's distance gets too large, grow the Map.
	// Defaults to 10.
	maxDistance int

	lock Spinlock
}

// Item represents an entry in the RHMap.
type Item struct {
	Key string
	Val interface{}

	Distance int // How far item is from its best position.
}

// New returns a new robinhood hashmap.
func New(capacity int) *Map {
	if capacity == 0 {
		capacity = defaultCapacity
	}
	return &Map{items: make([]Item, capacity)}
}

// Reset clears Map, where already allocated memory will be reused.
func (m *Map) Reset() {
	for i := range m.items {
		m.items[i] = Item{}
	}

	m.count = 0
}

func (m *Map) Get(key string) (val interface{}, err error) {
	size := int64(len(m.items))
	indexes := hashChain(size, key)
	for _, index := range indexes {
		if m.probeGet(key, index) {
			return m.items[index].Val, nil
		}
	}

	_, val, err = m.linearProbingGet(key, indexes[0])
	return val, err
}

func (m *Map) SyncGet(key string) (val interface{}, err error) {
	size := int64(len(m.items))
	indexes := hashChain(size, key)
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, index := range indexes {
		if m.probeGet(key, index) {
			return m.items[index].Val, nil
		}
	}

	_, val, err = m.linearProbingGet(key, indexes[0])
	return val, err
}

func (m *Map) Set(key string, val interface{}) error {
	size := int64(len(m.items))
	indexes := hashChain(size, key)
	for _, index := range indexes {
		if m.probeEmplace(key, val, index) {
			return nil
		}
	}

	return m.linearProbeEmplace(key, val, indexes[0])
}

func (m *Map) SyncSet(key string, val interface{}) error {
	size := int64(len(m.items))
	indexes := hashChain(size, key)
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, index := range indexes {
		if m.probeEmplace(key, val, index) {
			return nil
		}
	}

	return m.linearProbeEmplace(key, val, indexes[0])
}

func (m *Map) Del(key string) {
	size := int64(len(m.items))
	indexes := hashChain(size, key)
	for _, index := range indexes {
		if m.probeDel(key, index) {
			m.count--
			return
		}
	}

	idx, _, err := m.linearProbingGet(key, indexes[0])
	if !errors.Is(err, ErrNotFound) {
		m.count--
		m.items[idx].Key = "" // mark as deleted.
	}
}

func (m *Map) SyncDel(key string) {
	size := int64(len(m.items))
	indexes := hashChain(size, key)
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, index := range indexes {
		if m.probeDel(key, index) {
			m.count--
			return
		}
	}

	idx, _, err := m.linearProbingGet(key, indexes[0])
	if !errors.Is(err, ErrNotFound) {
		m.count--
		m.items[idx].Key = "" // mark as deleted.
	}
}

func (m *Map) probeGet(key string, index int64) bool {
	item := &m.items[index]
	return key == item.Key
}

// Get retrieves the val for a given key.
func (m *Map) linearProbingGet(key string, hint int64) (idx int64, val interface{}, err error) {
	num := int64(len(m.items))
	idxStart := hint

	for {
		e := &m.items[hint]
		if e.Key == "" {
			return 0, nil, ErrNotFound
		}

		if e.Key == key {
			return hint, e.Val, nil
		}

		hint++
		if hint >= num {
			hint = 0
		}

		if hint == idxStart { // Went all the way around.
			return 0, nil, ErrNotFound
		}
	}
}

func (m *Map) probeEmplace(key string, val interface{}, index int64) bool {
	item := &m.items[index]
	switch item.Key {
	case "":
		m.count++
		fallthrough
	case key:
		item.Key = key
		item.Val = val
		item.Distance = 0
		return true
	default:
		return false
	}
}

// linearProbeEmplace inserts or updates a key/val into the Map.
func (m *Map) linearProbeEmplace(key string, val interface{}, idx int64) error {
	num := int64(len(m.items))
	idxStart := idx

	incoming := Item{key, val, 0}

	for {
		e := &m.items[idx]
		if e.Key == "" {
			m.items[idx] = incoming
			m.count++
			return nil
		}

		if e.Key == incoming.Key {
			// NOTE: We keep the same key to allow advanced apps that
			// know that they're doing an update to avoid key alloc's.
			e.Val, e.Distance = incoming.Val, incoming.Distance

			return nil
		}

		// Swap if the incoming item is further from its best idx.
		if e.Distance < incoming.Distance {
			incoming, m.items[idx] = m.items[idx], incoming
		}

		incoming.Distance++ // One step further away from best idx.

		idx++
		if idx >= num {
			idx = 0
		}

		if incoming.Distance > m.maxDistance || idx == idxStart {
			m.Grow(m.count * 2)

			return m.Set(incoming.Key, incoming.Val)
		}
	}
}

func (m *Map) probeDel(key string, hint int64) bool {
	item := &m.items[hint]
	if item.Key == key {
		item.Key = "" // mark as deleted.
		return true
	}
	return false
}

// CopyTo copies key/val's to the dest Map.
func (m *Map) CopyTo(dest *Map) {
	m.Visit(func(k string, v interface{}) bool { dest.Set(k, v); return true })
}

// Visit invokes the callback on key/val. The callback can return
// false to exit the visitation early.
func (m *Map) Visit(callback func(k string, v interface{}) (keepGoing bool)) {
	for i := range m.items {
		e := &m.items[i]
		if e.Key != "" {
			if !callback(e.Key, e.Val) {
				return
			}
		}
	}
}

// Grow resizes Map size.
func (m *Map) Grow(newSize int) {
	grow := New(newSize)
	m.CopyTo(grow)
	m.Reset()
	m.items = grow.items
	m.count = grow.count
}
