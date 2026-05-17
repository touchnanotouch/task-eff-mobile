package model

import (
	"testing"
	"time"
)

func TestParseMonth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "valid month",
			input:   "07-2025",
			want:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "january",
			input:   "01-2026",
			want:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "december",
			input:   "12-2024",
			want:    time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid month 13",
			input:   "13-2025",
			wantErr: true,
		},
		{
			name:    "invalid month 00",
			input:   "00-2025",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "wrong separator",
			input:   "07/2025",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMonth(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMonth() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("ParseMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatMonth(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "july 2025",
			input: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			want:  "07-2025",
		},
		{
			name:  "january 2026",
			input: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			want:  "01-2026",
		},
		{
			name:  "december 2024",
			input: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			want:  "12-2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMonth(tt.input)
			if got != tt.want {
				t.Errorf("FormatMonth() = %s, want %s", got, tt.want)
			}
		})
	}
}
