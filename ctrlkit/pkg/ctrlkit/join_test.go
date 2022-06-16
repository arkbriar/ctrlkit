package ctrlkit

import "testing"

func Test_Join(t *testing.T) {
	if Join(Nop, Nop).Description() != "Join(Nop, Nop)" {
		t.Fatal("description of join is not correct")
	}
}

func Test_JoinInParallel(t *testing.T) {
	if JoinInParallel(Nop) != Nop {
		t.Fatal("one join should be optimized")
	}

	if JoinInParallel(Nop, Nop).Description() != "ParallelJoin(Nop, Nop)" {
		t.Fatal("description of parallel join is not correct")
	}
}
