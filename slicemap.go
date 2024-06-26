package slicemap

import (
	"slices"
	"sort"

	"github.com/goupdate/deadlock"
	"golang.org/x/exp/constraints"
)

// SliceMap is a map of slices of ordered values
type SliceMap[K constraints.Ordered, V constraints.Ordered] struct {
	deadlock.RWMutex
	data map[K]*[]V
}

// NewSliceMap creates a new SliceMap
func NewSliceMap[K, V constraints.Ordered]() *SliceMap[K, V] {
	return &SliceMap[K, V]{
		data: make(map[K]*[]V),
	}
}

// Add adds a value to the slice associated with the given key
func (sm *SliceMap[K, V]) Add(key K, value V) {
	sm.Lock()
	defer sm.Unlock()

	if slice, ok := sm.data[key]; ok {
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
		sm.data[key] = &[]V{value}
	}
}

// Delete removes a value from the slice associated with the given key
func (sm *SliceMap[K, V]) Delete(key K, value V) {
	sm.Lock()
	defer sm.Unlock()

	if slice, ok := sm.data[key]; ok {
		if len(*slice) > 0 {
			if value == (*slice)[0] {
				// Remove from the beginning
				*slice = (*slice)[1:]
				if len(*slice) == 0 {
					delete(sm.data, key)
				}
				return
			} else if value == (*slice)[len(*slice)-1] {
				// Remove from the end
				*slice = (*slice)[:len(*slice)-1]
				if len(*slice) == 0 {
					delete(sm.data, key)
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
				delete(sm.data, key)
			}
		}
	}
}

// DeleteKey removes the key and its associated slice from the map
func (sm *SliceMap[K, V]) DeleteKey(key K) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.data, key)
}

// Count returns the total number of elements in all slices
func (sm *SliceMap[K, V]) Count() int64 {
	sm.RLock()
	defer sm.RUnlock()

	var count int64
	for _, slice := range sm.data {
		count += int64(len(*slice))
	}
	return count
}

// GetKey returns the slice associated with the given key
func (sm *SliceMap[K, V]) GetKey(key K) *[]V {
	sm.RLock()
	defer sm.RUnlock()

	if slice, ok := sm.data[key]; ok {
		return slice
	}
	return nil
}

// IterateValues iterates over all key-value pairs in the map
func (sm *SliceMap[K, V]) IterateValues(f func(K, V) bool) {
	sm.RLock()
	defer sm.RUnlock()

	for k, slice := range sm.data {
		for _, v := range *slice {
			if !f(k, v) {
				return
			}
		}
	}
}

// IterateKeys iterates over all keys in the map
func (sm *SliceMap[K, V]) IterateKeys(f func(K) bool) {
	sm.RLock()
	defer sm.RUnlock()

	for k := range sm.data {
		if !f(k) {
			return
		}
	}
}

// Exist checks if the value v exists for the key k
func (sm *SliceMap[K, V]) Exist(key K, value V) bool {
	sm.RLock()
	defer sm.RUnlock()

	if slice, ok := sm.data[key]; ok {
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

	if slice, was := sm.data[key]; !was {
		// Если ключа нет, просто копируем values
		ns := make([]V, len(values))
		copy(ns, values)
		sm.data[key] = &ns
	} else {
		//sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		// Если ключ есть, объединяем новые и старые значения, сохраняя уникальность и порядок
		*slice = mergeUniqueSorted(*slice, values)
	}
}

// AddSlice adds multiple values to the slice associated with the given key
func (sm *SliceMap[K, V]) AddSliceSortCompact(key K, values []V) {
	sm.Lock()
	defer sm.Unlock()

	slices.Sort(values)
	values = slices.Compact(values)

	if slice, was := sm.data[key]; !was {
		// Если ключа нет, просто копируем values
		ns := make([]V, len(values))
		copy(ns, values)
		sm.data[key] = &ns
	} else {
		//sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		// Если ключ есть, объединяем новые и старые значения, сохраняя уникальность и порядок
		*slice = mergeUniqueSorted(*slice, values)
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

	// Добавляем оставшиеся элементы
	for ; i < len(a); i++ {
		result = append(result, a[i])
	}
	for ; j < len(b); j++ {
		result = append(result, b[j])
	}
	return result
}

// GetStorage returns a reference to the internal map
// dont forget use RLock, RUnlock !
func (sm *SliceMap[K, V]) GetStorageNotLocked() (*map[K]*[]V, *deadlock.RWMutex) {
	// Возвращаем указатель на мьютекс и данные
	return &sm.data, &sm.RWMutex
}

// GetMutex returns a reference to the internal mutex
func (sm *SliceMap[K, V]) GetMutex() *deadlock.RWMutex {
	return &sm.RWMutex
}
