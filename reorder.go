package main

// buildOrderIndices returns the 0-based page indices in the order they should
// be merged so that duplex printing (short-edge) yields fronts 1..S and backs S+1..N.
// If n is odd, we conceptually pad to the next even number (the caller may append a blank).
func buildOrderIndices(n int) []int {
	// Pad for odd page count to keep pairs intact.
	if n%2 != 0 {
		n++
	}
	s := n / 2
	order := make([]int, 0, n)
	for i := 0; i < s; i++ {
		order = append(order, i)   // front i+1
		order = append(order, s+i) // back  s+i+1
	}
	return order
}
