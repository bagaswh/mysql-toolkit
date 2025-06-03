package queryfingerprint

import "testing"

func Test_Fingerprint(t *testing.T) {
	testCases := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "simple select",
			sql:      "select * from users",
			expected: "select * from users",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Fingerprint(tc.sql)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if actual != tc.expected {
				t.Errorf("expected: %s, actual: %s", tc.expected, actual)
			}
		})
	}
}
