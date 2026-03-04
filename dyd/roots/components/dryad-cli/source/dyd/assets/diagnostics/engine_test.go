package diagnostics

import "testing"

func TestCompileMetricsSampleEvery_RoundingContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		percent  *float64
		expected uint64
	}{
		{
			name:     "nil defaults to 100 percent",
			percent:  nil,
			expected: 1,
		},
		{
			name:     "100 percent maps to every call",
			percent:  floatRef(100),
			expected: 1,
		},
		{
			name:     "50 percent maps to one in two",
			percent:  floatRef(50),
			expected: 2,
		},
		{
			name:     "30 percent rounds to one in four",
			percent:  floatRef(30),
			expected: 4,
		},
		{
			name:     "25 percent maps to one in four",
			percent:  floatRef(25),
			expected: 4,
		},
		{
			name:     "12.5 percent maps to one in eight",
			percent:  floatRef(12.5),
			expected: 8,
		},
		{
			name:     "1 percent rounds to one in one hundred twenty eight",
			percent:  floatRef(1),
			expected: 128,
		},
		{
			name:     "75 percent tie favors lower capture rate",
			percent:  floatRef(75),
			expected: 2,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			every, err := compileMetricsSampleEvery(test.percent)
			if err != nil {
				t.Fatalf("compileMetricsSampleEvery returned error: %v", err)
			}
			if every != test.expected {
				t.Fatalf("expected sampleEvery=%d, got %d", test.expected, every)
			}
		})
	}
}

func TestCompileMetricsSampleEvery_RejectsOutOfRange(t *testing.T) {
	t.Parallel()

	tests := []float64{
		0,
		-1,
		100.0001,
	}

	for _, percent := range tests {
		percent := percent
		t.Run("percent", func(t *testing.T) {
			t.Parallel()

			_, err := compileMetricsSampleEvery(&percent)
			if err == nil {
				t.Fatalf("expected error for sample_percent=%v", percent)
			}
		})
	}
}
