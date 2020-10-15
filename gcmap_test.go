package worker

import (
	"testing"
	"time"
)

func TestStorage_Range(t *testing.T) {
	st := NewStorage()
	st.Store("a", 1)
	st.Store("b", 1)
	st.Store("c", 1)
	st.Store("d", 1)

	var counter int
	st.Range(func(k, v interface{}) bool {
		counter++
		return true
	})

	if counter != 4 {
		t.Fatal("range had invalid number of iterations")
	}
}

func TestStorage_StoreLoad(t *testing.T) {
	st := NewStorage()
	st.Store("a", "b")
	value, found := st.Load("a")
	if !found || value != "b" {
		t.Fatal("loaded value is not set or invalid")
	}
}

func TestStorage_StoreOrUpdate(t *testing.T) {
	type testStruct struct {
		A string
		B string
	}

	st := NewStorage()
	st.StoreOrUpdate("key", testStruct{A: "a"}, nil)
	st.StoreOrUpdate("key", testStruct{A: "", B: "b"}, func(old, new interface{}) interface{} {
		return testStruct{
			A: old.(testStruct).A,
			B: new.(testStruct).B,
		}
	})
	current, found := st.Load("key")
	if !found || current == nil {
		t.Error("failed to to find provided key in a storage")
	}
	// spew.Dump(current)
	if current.(testStruct).A != "a" || current.(testStruct).B != "b" {
		t.Error("value in a storage was incorrectly updated")
	}
}

func TestStorage_Delete(t *testing.T) {
	st := NewStorage()
	st.Store("a", "b")
	st.Delete("a")
	_, found := st.Load("a")
	if found {
		t.Fatal("loaded previously deleted value")
	}
}

func TestStorage_GC(t *testing.T) {
	st := NewStorage(
		WithGCInterval(time.Millisecond),
		WithEntryTTL(time.Millisecond),
	)
	st.Store("a", "1")
	st.Store("b", "1")
	st.Store("c", "1")
	st.Store("d", "1")
	time.Sleep(time.Millisecond * 5)

	var counter int
	st.Range(func(k, v interface{}) bool {
		counter++
		return true
	})

	if counter != 0 {
		t.Errorf("storage supposed to be empty after GC, but it was of %d size", counter)
	}
}

func BenchmarkStore(b *testing.B) {
	st := NewStorage(WithGCInterval(0))
	for i := 0; i < b.N; i++ {
		st.Store(i, i)
	}
}

func BenchmarkLoadAndStore(b *testing.B) {
	st := NewStorage(WithGCInterval(0))
	for i := 0; i < b.N; i++ {
		st.Store(i, i)
		st.Load(i)
	}
}
