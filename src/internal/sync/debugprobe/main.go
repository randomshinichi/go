package main

import "internal/sync"

type setting struct {
	value string
}

func main() {
	for n := 0; n < 1000; n++ {
		var m sync.HashTrieMap[any, any]
		value := &setting{value: "#failfipscast"}
		got, loaded := m.LoadOrStore(value.value, value)
		if loaded || got != value {
			println("unexpected map result", n, loaded)
			panic("unexpected map result")
		}
	}
	println("HashTrieMap any/any zero-state probe passed")
}
