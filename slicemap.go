package slicemap

import (
	"slices"
	"sort"
	"sync"

	"golang.org/x/exp/constraints"
)

// SliceMap is a map of slices of ordered values
type SliceMap[K constraints.Ordered, V constraints.Ordered] struct {
	sync.RWMutex
	data sync.Map
}

// NewSliceMap creates a new SliceMap
func NewSliceMap[K, V constraints.Ordered]() *SliceMap[K, V] {
	return &SliceMap[K, V]{}
}

// Add adds a value to the slice associated with the given key
func (sm *SliceMap[K, V]) Add(key K, value V) {
	sm.Lock()
	defer sm.Unlock()

	if slice_, ok := sm.data.Load(key); ok {
		slice := slice_.(*[]V)
		if len(*slice) > 0 {
			if value < (*slice)[0] {
				// Insert at the beginning
				*slice = append([]V{value}, *slice...)
				return
			} else if value > (*slice)[len(*slice)-1] {
				// Insert at the end
				*slice = append(*slice, value)
				return
			}

			if (*slice)[0] == value || (*slice)[len(*slice)-1] == value {
				return // Value already exists
			}
		}

		// Binary search to find the insertion point
		i := sort.Search(len(*slice), func(i int) bool { return (*slice)[i] >= value })
		if i < len(*slice) && (*slice)[i] == value {
			return // Value already exists
		}
		// Insert value at the index found
		*slice = append(*slice, value)
		copy((*slice)[i+1:], (*slice)[i:])
		(*slice)[i] = value
	} else {
		// Create a new slice and add value
		sm.data.Store(key, &[]V{value})
	}
}

// Delete removes a value from the slice associated with the given key
func (sm *SliceMap[K, V]) Delete(key K, value V) {
	sm.Lock()
	defer sm.Unlock()

	if slice_, ok := sm.data.Load(key); ok {
		slice := slice_.(*[]V)
		if len(*slice) > 0 {
			if value == (*slice)[0] {
				// Remove from the beginning
				*slice = (*slice)[1:]
				if len(*slice) == 0 {
					sm.data.Delete(key)
				}
				return
			} else if value == (*slice)[len(*slice)-1] {
				// Remove from the end
				*slice = (*slice)[:len(*slice)-1]
				if len(*slice) == 0 {
					sm.data.Delete(key)
				}
				return
			} else if value < (*slice)[0] || value > (*slice)[len(*slice)-1] {
				return // Value is out of the range of the slice
			}
		}

		i := sort.Search(len(*slice), func(i int) bool { return (*slice)[i] >= value })
		if i < len(*slice) && (*slice)[i] == value {
			// Remove the element at index i
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			if len(*slice) == 0 {
				sm.data.Delete(key)
			}
		}
	}
}

// DeleteKey removes the key and its associated slice from the map
func (sm *SliceMap[K, V]) DeleteKey(key K) {
	sm.data.Delete(key)
}

// Count returns the total number of elements in all slices
func (sm *SliceMap[K, V]) Count() int64 {
	var count int64
	sm.data.Range(func(k, v interface{}) bool {
		sm.RLock()
		defer sm.RUnlock()

		slice := v.(*[]V)
		count += int64(len(*slice))
		return true
	})
	return count
}

// GetKey returns COPY of the slice associated with the given key
func (sm *SliceMap[K, V]) GetKey(key K) *[]V {
	if slice_, ok := sm.data.Load(key); ok {
		sm.RLock()
		defer sm.RUnlock()

		slice := slice_.(*[]V)
		cpy := slices.Clone(*slice)
		return &cpy
	}
	return nil
}

// IterateValues iterates over all key-value pairs in the map
func (sm *SliceMap[K, V]) IterateValues(f func(K, V) bool) {
	sm.data.Range(func(k, slice_ interface{}) bool {
		sm.RLock()
		slice := slice_.(*[]V)
		cpy := slices.Clone(*slice)
		sm.RUnlock()

		for _, v := range cpy {
			if !f(k.(K), v) {
				return false
			}
		}
		return true
	})
}

// IterateKeys iterates over all keys in the map
func (sm *SliceMap[K, V]) IterateKeys(f func(K) bool) {
	sm.data.Range(func(k, _ interface{}) bool {
		if !f(k.(K)) {
			return false
		}
		return true
	})
}

// Exist checks if the value v exists for the key k
func (sm *SliceMap[K, V]) Exist(key K, value V) bool {
	if slice_, ok := sm.data.Load(key); ok {
		sm.RLock()
		defer sm.RUnlock()

		slice := slice_.(*[]V)
		if len(*slice) == 0 {
			return false
		}
		if value < (*slice)[0] || value > (*slice)[len(*slice)-1] {
			return false // Value is out of the range
		}
		i := sort.Search(len(*slice), func(i int) bool { return (*slice)[i] >= value })
		return i < len(*slice) && (*slice)[i] == value
	}
	return false
}

// AddSlice adds multiple values to the slice associated with the given key
func (sm *SliceMap[K, V]) AddSlice(key K, values []V) {
	sm.Lock()
	defer sm.Unlock()

	slices.Sort(values)
	values = slices.Compact(values)

	slice_, was := sm.data.Load(key)

	if !was {
		// Если ключа нет, просто копируем values
		ns := make([]V, len(values))
		copy(ns, values)
		sm.data.Store(key, &ns)
	} else {
		//sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		// Если ключ есть, объединяем новые и старые значения, сохраняя уникальность и порядок
		slice := slice_.(*[]V)
		*slice = mergeUniqueSorted(*slice, values)
		sm.data.Store(key, slice)
	}
}

// mergeUniqueSorted объединяет два отсортированных слайса в один уникальный отсортированный слайс
func mergeUniqueSorted[V constraints.Ordered](a, b []V) []V {
	result := make([]V, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			result = append(result, a[i])
			i++
		} else if a[i] > b[j] {
			result = append(result, b[j])
			j++
		} else {
			result = append(result, a[i]) // Оба значения равны, добавляем один раз
			i++
			j++
		}
	}

	for ; i < len(a); i++ {
		result = append(result, a[i])
	}
	for ; j < len(b); j++ {
		result = append(result, b[j])
	}
	return result
}
