package simple

func putBytes(b []byte, bs ...[]byte) []byte {
	i := 0
	for _, b2 := range bs {
		copy(b[i:], b2)
		i += len(b2) - 1
	}
	return b
}
