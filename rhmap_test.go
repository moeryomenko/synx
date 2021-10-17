package synx

import (
	"testing"
)

func TestMap(t *testing.T) {
	r := New(100)
	ops := 0

	g := map[string]string{}

	get := func(k string) {
		ops++

		v, err := r.Get(k)
		rv := ""
		if err != ErrNotFound {
			rv = v.(string)
		}

		gv := g[k]

		if rv != gv {
			t.Fatalf("ops: %d, get different values for key %s, expected: %s, but given %s",
				ops, k, gv, rv)
		}
		if r.Count != len(g) {
			t.Fatalf("ops: %d, get different counts", ops)
		}
	}

	set := func(k, v string) {
		ops++

		err := r.Set(k, v)
		g[k] = v

		if err != nil {
			t.Fatalf("ops: %d, set err", ops)
		}
		if r.Count != len(g) {
			t.Fatalf("ops: %d, set different counts", ops)
		}
	}

	del := func(k string) {
		ops++

		r.Del(k)
		delete(g, k)

		if r.Count != len(g) {
			t.Fatalf("ops: %d, del different counts", ops)
		}
	}

	// ------------------------------------------

	get("not a key")
	get("nothing there")

	set("a", "A")
	get("a")
	get("b")

	set("a", "AA")
	get("a")
	get("b")

	set("b", "B")
	get("a")
	get("b")
	get("c")

	get("not a key")
	get("nothing there")

	del("a")
	get("a")
	get("b")
	get("c")

	del("a")
	get("a")
	get("b")
	get("c")

	del("b")
	get("a")
	get("b")
	get("c")

	set("a", "A")
	set("b", "B")
	set("c", "C")
	set("d", "D")
	set("e", "E")
	set("f", "F")
	set("a1", "")
	set("b1", "")
	set("c1", "C1")
	set("d1", "D1")
	set("e1", "E1")
	set("f1", "F1")
	set("a11", "A11")
	set("b11", "B11")
	set("c11", "C11")
	set("d11", "D11")
	set("e11", "E11")
	set("f11", "F11") // 18 entries.

	get("a")
	get("b")
	get("c")
	get("d")
	get("e")
	get("f")

	get("not a key")
	get("nothing there")
}
