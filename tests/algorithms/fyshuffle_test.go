package	test

import (
	. "fortuna/core/alg"
	"testing"
)

func TestGetUniqueRandomNumber(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		count   int
		max     int
		wantErr bool
	}{
		{
			name:    "Normal case",
			seed:    42,
			count:   5,
			max:     10,
			wantErr: false,
		},
		{
			name:    "Count equals max + 1",
			seed:    100,
			count:   11,
			max:     10,
			wantErr: false,
		},
		{
			name:    "Count greater than max + 1",
			seed:    50,
			count:   12,
			max:     10,
			wantErr: true,
		},
		{
			name:    "Zero count",
			seed:    0,
			count:   0,
			max:     10,
			wantErr: false,
		},
		{
			name:    "Max zero with count one",
			seed:    1,
			count:   1,
			max:     0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUniqueRandomNumber(tt.seed, tt.count, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUniqueRandomNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.count {
					t.Errorf("Expected %d numbers, got %d", tt.count, len(got))
				}
				numMap := make(map[int]bool)
				for _, num := range got {
					if num < 0 || num > tt.max {
						t.Errorf("Number %d out of range [0, %d]", num, tt.max)
					}
					if numMap[num] {
						t.Errorf("Duplicate number found: %d", num)
					}
					numMap[num] = true
				}
			}
		})
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name    string
		seed    int64
		numbers []int
		count   int
		max     int
		want    bool
		wantErr bool
	}{
		{
			name:    "Valid verification",
			seed:    42,
			count:   5,
			max:     10,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Invalid numbers (wrong seed)",
			seed:    42,
			count:   5,
			max:     10,
			numbers: []int{1, 2, 3, 4, 5}, // Assume these are incorrect for seed 42
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid numbers (duplicates)",
			seed:    42,
			count:   5,
			max:     10,
			numbers: []int{1, 2, 2, 4, 5},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid count",
			seed:    42,
			count:   5,
			max:     10,
			numbers: []int{1, 2, 3}, // Less than count
			want:    false,
			wantErr: false,
		},
		{
			name:    "Valid verification with count equals max+1",
			seed:    100,
			count:   11,
			max:     10,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Empty numbers with zero count",
			seed:    0,
			count:   0,
			max:     10,
			numbers: []int{},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Mismatch in count",
			seed:    1,
			count:   1,
			max:     0,
			numbers: []int{1}, // Should be out of range
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Verify(tt.seed, tt.numbers, tt.count, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
