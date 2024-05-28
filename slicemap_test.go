package slicemap

import (
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
