package service

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Қ", "к"},
		{"қ", "к"},
		{"Ң", "н"},
		{"ң", "н"},
		{"Ү", "у"},
		{"ү", "у"},
		{"Ө", "о"},
		{"ө", "о"},
		{"Ғ", "г"},
		{"ғ", "г"},
		{"Ұ", "у"},
		{"ұ", "у"},
		{"Ә", "а"},
		{"ә", "а"},
		{"І", "и"},
		{"і", "и"},
		{"Һ", "х"},
		{"һ", "х"},
		// mixed Kazakh+Russian string
		{"Қуаныш", "куаныш"},
		{"Толеу", "толеу"},
		// pure Russian — no-op except lowercase
		{"Иванов", "иванов"},
		{"ПЕТРОВ", "петров"},
		// already lower — unchanged
		{"алибек", "алибек"},
	}

	for _, tc := range tests {
		got := normalizeName(tc.input)
		if got != tc.want {
			t.Errorf("normalizeName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
