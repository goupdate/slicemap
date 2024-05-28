package slicemap

import (
	"sort"
	"testing"
)

func TestSliceMapOperations(t *testing.T) {
	sm := NewSliceMap[int, int]()
	// Test adding elements
	sm.Add(1, 10)
	sm.Add(1, 20)
	sm.Add(2, 10)
	if len(*sm.GetKey(1)) != 2 {
		t.Errorf("Expected 2 elements for key 1, got %d", len(*sm.GetKey(1)))
	}

	if !sm.Exist(1, 20) {
		t.Errorf("Value not exist but should")
	}

	// Test deleting an element
	sm.Delete(1, 10)
	if len(*sm.GetKey(1)) != 1 {
		t.Errorf("Expected 1 element for key 1 after deletion, got %d", len(*sm.GetKey(1)))
	}

	if sm.Exist(1, 10) {
		t.Errorf("Value exist but should not")
	}

	// Ensure key is removed if slice is empty
	sm.Delete(1, 20)
	if sm.GetKey(1) != nil {
		t.Errorf("Expected nil for key 1 after deleting all elements, got %v", sm.GetKey(1))
	}

	// Test deleting a key directly
	sm.Add(3, 30)
	sm.DeleteKey(3)
	if sm.GetKey(3) != nil {
		t.Errorf("Expected nil for key 3 after deleting the key, got %v", sm.GetKey(3))
	}
}

func TestConcurrency(t *testing.T) {
	sm := NewSliceMap[int, int]()
	// Run Add and Delete in parallel
	go func() {
		for i := 0; i < 1000; i++ {
			sm.Add(1, i)
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			sm.Delete(1, i)
		}
	}()

	// Allow some time for operations to complete
	t.Parallel()
}

func BenchmarkAddDelete(b *testing.B) {
	sm := NewSliceMap[int, int]()
	for i := 0; i < b.N; i++ {
		sm.Add(1, i)
		sm.Delete(1, i)
	}
}

func BenchmarkAdd(b *testing.B) {
	sm := NewSliceMap[int, int]()
	for i := 0; i < b.N; i++ {
		sm.Add(1, i)
	}
}

func BenchmarkDelete(b *testing.B) {
	sm := NewSliceMap[int, int]()
	for i := 0; i < b.N; i++ {
		sm.Add(1, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Delete(1, i)
	}
}

func TestAddSlice(t *testing.T) {
	sm := NewSliceMap[int, int]()

	// Тестирование добавления нового слайса
	sm.AddSlice(1, []int{5, 3, 8})
	if slice := sm.GetKey(1); slice == nil || len(*slice) != 3 {
		t.Fatalf("Expected 3 elements for key 1, got %d", len(*slice))
	}
	if !sort.IntsAreSorted(*sm.GetKey(1)) {
		t.Errorf("Slice for key 1 should be sorted")
	}

	// Тестирование добавления элементов с дубликатами
	sm.AddSlice(1, []int{3, 7, 2, 2})
	if slice := sm.GetKey(1); slice == nil || len(*slice) != 5 {
		t.Fatalf("Expected 5 unique elements for key 1, got %d, %v", len(*slice), *slice)
	}
	expectedSlice := []int{2, 3, 5, 7, 8}
	for i, v := range *sm.GetKey(1) {
		if v != expectedSlice[i] {
			t.Errorf("Expected element %d at index %d, got %d", expectedSlice[i], i, v)
		}
	}

	// Тестирование добавления слайса к новому ключу
	sm.AddSlice(2, []int{22, 6, 1})
	if slice := sm.GetKey(2); slice == nil || len(*slice) != 3 {
		t.Fatalf("Expected 3 elements for key 2, got %d", len(*slice))
	}
	if !sort.IntsAreSorted(*sm.GetKey(2)) {
		t.Errorf("Slice for key 2 should be sorted")
	}
}
