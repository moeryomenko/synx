package synx

import "errors"

var (
	ErrMapFull  = errors.New("map is full")
	ErrNotFound = errors.New("key not found")
)

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
	// Items are the slots of the hashmap for items.
	Items []Item

	// Number of keys in the Map.
	Count int

	lock Spinlock
}

// Item represents an entry in the RHMap.
type Item struct {
	Key string
	Val interface{}

	Distance int // How far item is from its best position.
}

// New returns a new robinhood hashmap.
func New(size int) *Map {
	return &Map{Items: make([]Item, size)}
}

// Reset clears Map, where already allocated memory will be reused.
func (m *Map) Reset() {
	for i := range m.Items {
		m.Items[i] = Item{}
	}

	m.Count = 0
}

func (m *Map) Get(key string) (val interface{}, err error) {
	size := int64(len(m.Items))
	indexes := hashChain(size, key)
	for _, index := range indexes {
		if m.probeGet(key, index) {
			return m.Items[index].Val, nil
		}
	}

	_, val, err = m.linearProbingGet(key, indexes[0])
	return val, err
}

func (m *Map) SyncGet(key string) (val interface{}, err error) {
	size := int64(len(m.Items))
	indexes := hashChain(size, key)
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, index := range indexes {
		if m.probeGet(key, index) {
			return m.Items[index].Val, nil
		}
	}

	_, val, err = m.linearProbingGet(key, indexes[0])
	return val, err
}

func (m *Map) Set(key string, val interface{}) error {
	size := int64(len(m.Items))
	indexes := hashChain(size, key)
	for _, index := range indexes {
		if m.probeEmplace(key, val, index) {
			return nil
		}
	}

	return m.linearProbeEmplace(key, val, indexes[0])
}

func (m *Map) SyncSet(key string, val interface{}) error {
	size := int64(len(m.Items))
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
	size := int64(len(m.Items))
	indexes := hashChain(size, key)
	for _, index := range indexes {
		if m.probeDel(key, index) {
			m.Count--
			return
		}
	}

	idx, _, err := m.linearProbingGet(key, indexes[0])
	if !errors.Is(err, ErrNotFound) {
		m.Count--
		m.Items[idx].Key = "" // mark as deleted.
	}
}

func (m *Map) SyncDel(key string) {
	size := int64(len(m.Items))
	indexes := hashChain(size, key)
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, index := range indexes {
		if m.probeDel(key, index) {
			m.Count--
			return
		}
	}

	idx, _, err := m.linearProbingGet(key, indexes[0])
	if !errors.Is(err, ErrNotFound) {
		m.Count--
		m.Items[idx].Key = "" // mark as deleted.
	}
}

func (m *Map) probeGet(key string, index int64) bool {
	item := &m.Items[index]
	return key == item.Key
}

// Get retrieves the val for a given key.
func (m *Map) linearProbingGet(key string, hint int64) (idx int64, val interface{}, err error) {
	num := int64(len(m.Items))
	idxStart := hint

	for {
		e := &m.Items[hint]
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
	item := &m.Items[index]
	switch item.Key {
	case "":
		m.Count++
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
	num := int64(len(m.Items))
	idxStart := idx

	incoming := Item{key, val, 0}

	for {
		e := &m.Items[idx]
		if e.Key == "" {
			m.Items[idx] = incoming
			m.Count++
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
			incoming, m.Items[idx] = m.Items[idx], incoming
		}

		incoming.Distance++ // One step further away from best idx.

		idx++
		if idx >= num {
			idx = 0
		}

		if idx == idxStart {
			return ErrMapFull
		}
	}
}

func (m *Map) probeDel(key string, hint int64) bool {
	item := &m.Items[hint]
	if item.Key == key {
		item.Key = "" // mark as deleted.
		return true
	}
	return false
}
