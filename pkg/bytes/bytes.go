package bytes

func ToLowerInPlace(b []byte) []byte {
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] |= 0x20
		}
	}
	return b
}

func ToUpperInPlace(b []byte) []byte {
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] &^= 0x20
		}
	}
	return b
}

func PutBytes(dst []byte, bs ...[]byte) (int, []byte) {
	off := 0
	for _, b := range bs {
		off += copy(dst[off:], b)
		if len(b)+off > cap(dst) {
			return off, dst
		}
	}
	return off, dst
}
