package snowflake

import (
	"testing"
	"time"
)

var (
	testInstanceID = uint64(1) | uint64(1)<<9
	testStartTime  = time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func BenchmarkSnowflake_NewID(b *testing.B) {
	s := New(testStartTime, testInstanceID)
	defer s.Stop()
	for n := 0; n < b.N; n++ {
		s.NewID()
	}
}

func TestSnowflake_NewID(t *testing.T) {
	s := New(testStartTime, testInstanceID)
	defer s.Stop()
	var id uint64
	for i := 0; i < 10; i++ {
		id = s.NewID()
	}
	if s.Count() != 10 {
		t.Fatal("bad number of count")
	}
	t.Logf("ins: %b, seq: %b, mask: %b, id: %b", testInstanceID, id&uint12Mask, uint12Mask, id)
	t.Logf("ins: %x, seq: %x, mask: %x, id: %x", testInstanceID, id&uint12Mask, uint12Mask, id)
	t.Logf("ins: %d, seq: %d, mask: %d, id: %d", testInstanceID, id&uint12Mask, uint12Mask, id)
	if id&uint12Mask != 9 {
		t.Fatal("bad sequence id")
	}
	if (id>>12)&uint10Mask != testInstanceID {
		t.Fatal("bad instance id")
	}
	if time.Since(testStartTime)/time.Second != time.Duration(id>>22)*time.Millisecond/time.Second {
		t.Fatal("bad timestamp")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		} else {
			t.Logf("%v", r)
		}
	}()
	s2 := New(testStartTime, testInstanceID)
	s2.Stop()
	s2.NewID()
}

func TestSnowflake_CheckDup(t *testing.T) {
	s := New(testStartTime, testInstanceID)
	defer s.Stop()

	out := map[uint64]bool{}

	for i := 0; i < 100000; i++ {
		id := s.NewID()
		if out[id] {
			t.Fatal("duplicated")
		}
		out[id] = true
	}
}
