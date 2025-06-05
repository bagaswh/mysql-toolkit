package bytes

import (
	"bytes"
	"testing"
)

func TestPutBytes(t *testing.T) {
	testCases := []struct {
		name        string
		createDst   func() []byte
		input       [][]byte
		expected    []byte
		copiedCount int
	}{
		{
			name: "empty slice",
			createDst: func() []byte {
				return make([]byte, 0)
			},
			input:       [][]byte{},
			expected:    []byte{},
			copiedCount: 0,
		},
		{
			name: "dst is larger than input",
			createDst: func() []byte {
				return make([]byte, 20)
			},
			input: [][]byte{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				{11, 12, 13},
			},
			expected: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
				11, 12, 13, 0, 0, 0, 0, 0, 0, 0,
			},
			copiedCount: 13,
		},
		{
			name: "dst is smaller than input",
			createDst: func() []byte {
				return make([]byte, 5)
			},
			input: [][]byte{
				{1, 2, 3, 4, 5},
				{6, 7, 8, 9},
			},
			expected: []byte{
				1, 2, 3, 4, 5,
			},
			copiedCount: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := tc.createDst()
			n, result := PutBytes(dst[:], tc.input...)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("PutBytes() = %q; want %q", dst, tc.expected)
			}
			if n != tc.copiedCount {
				t.Errorf("PutBytes() copied %d bytes; want %d", n, tc.copiedCount)
			}
		})
	}
}

// TestToLowerInPlace tests the ToLowerInPlace function.
func TestToLowerInPlace(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "all uppercase",
			input:    []byte("HELLO"),
			expected: []byte("hello"),
		},
		{
			name:     "all lowercase",
			input:    []byte("world"),
			expected: []byte("world"),
		},
		{
			name:     "mixed case",
			input:    []byte("GoLang"),
			expected: []byte("golang"),
		},
		{
			name:     "with numbers and symbols",
			input:    []byte("Hello 123!"),
			expected: []byte("hello 123!"),
		},
		{
			name:     "only numbers and symbols",
			input:    []byte("123 !@#"),
			expected: []byte("123 !@#"),
		},
		{
			name:     "already manipulated (should not double apply or corrupt)",
			input:    ToLowerInPlace([]byte("ALREADY LOWER")), // apply once
			expected: []byte("already lower"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a copy for in-place modification to avoid issues if tc.input is reused
			inputCopy := make([]byte, len(tc.input))
			copy(inputCopy, tc.input)

			result := ToLowerInPlace(inputCopy)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("ToLowerInPlace(%q) = %q; want %q", tc.input, result, tc.expected)
			}
			// Additionally check if the original slice reference was modified (as it should be for "InPlace")
			if len(tc.input) > 0 && !bytes.Equal(inputCopy, tc.expected) {
				t.Errorf("ToLowerInPlace did not modify the slice in place as expected. Got %q, want %q", inputCopy, tc.expected)
			}
		})
	}
}

// TestToUpperInPlace tests the ToUpperInPlace function.
func TestToUpperInPlace(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "all lowercase",
			input:    []byte("hello"),
			expected: []byte("HELLO"),
		},
		{
			name:     "all uppercase",
			input:    []byte("WORLD"),
			expected: []byte("WORLD"),
		},
		{
			name:     "mixed case",
			input:    []byte("GoLang"),
			expected: []byte("GOLANG"),
		},
		{
			name:     "with numbers and symbols",
			input:    []byte("Hello 123!"),
			expected: []byte("HELLO 123!"),
		},
		{
			name:     "only numbers and symbols",
			input:    []byte("123 !@#"),
			expected: []byte("123 !@#"),
		},
		{
			name:     "already manipulated (should not double apply or corrupt)",
			input:    ToUpperInPlace([]byte("already upper")), // apply once
			expected: []byte("ALREADY UPPER"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a copy for in-place modification
			inputCopy := make([]byte, len(tc.input))
			copy(inputCopy, tc.input)

			result := ToUpperInPlace(inputCopy)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("ToUpperInPlace(%q) = %q; want %q", tc.input, result, tc.expected)
			}
			// Additionally check if the original slice reference was modified
			if len(tc.input) > 0 && !bytes.Equal(inputCopy, tc.expected) {
				t.Errorf("ToUpperInPlace did not modify the slice in place as expected. Got %q, want %q", inputCopy, tc.expected)
			}
		})
	}
}

// --- Benchmarks ---

var benchData = []byte("TheQuickBrownFoxJumpsOverTheLazyDogAndTheQuickBrownFoxJumpsOverTheLazyDog")
var benchDataLower = bytes.ToLower(benchData)
var benchDataUpper = bytes.ToUpper(benchData)

// BenchmarkToLowerInPlace benchmarks the ToLowerInPlace function.
func BenchmarkToLowerInPlace(b *testing.B) {
	// Create a new slice for each iteration to ensure in-place modification doesn't affect subsequent runs.
	data := make([]byte, len(benchDataUpper))
	for i := 0; i < b.N; i++ {
		copy(data, benchDataUpper) // Reset data to uppercase for a fair ToLower test
		ToLowerInPlace(data)
	}
}

// BenchmarkStdLibToLower benchmarks the standard library bytes.ToLower function for comparison.
func BenchmarkStdLibToLower(b *testing.B) {
	data := make([]byte, len(benchDataUpper))
	for i := 0; i < b.N; i++ {
		copy(data, benchDataUpper)
		_ = bytes.ToLower(data) // bytes.ToLower returns a new slice, but we mimic in-place cost by copying first
	}
}

// BenchmarkToUpperInPlace benchmarks the ToUpperInPlace function.
func BenchmarkToUpperInPlace(b *testing.B) {
	data := make([]byte, len(benchDataLower))
	for i := 0; i < b.N; i++ {
		copy(data, benchDataLower) // Reset data to lowercase for a fair ToUpper test
		ToUpperInPlace(data)
	}
}

// BenchmarkStdLibToUpper benchmarks the standard library bytes.ToUpper function for comparison.
func BenchmarkStdLibToUpper(b *testing.B) {
	data := make([]byte, len(benchDataLower))
	for i := 0; i < b.N; i++ {
		copy(data, benchDataLower)
		_ = bytes.ToUpper(data) // bytes.ToUpper returns a new slice
	}
}
