package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{
			name:    "return nil when number is valid",
			number:  "12345678903",
			wantErr: false,
		},
		{
			name:    "return nil when number is valid",
			number:  "9278923470",
			wantErr: false,
		},
		{
			name:    "return error when number is invalid",
			number:  "49927398717",
			wantErr: true,
		},
		{
			name:    "return error when number is invalid",
			number:  "123456781231",
			wantErr: true,
		},
		{
			name:    "return error when number contains non-numeric characters",
			number:  "abcd1234",
			wantErr: true,
		},
		{
			name:    "return error when number is empty",
			number:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ValidateOrderNumber(%s): %s", tt.number, tt.name), func(t *testing.T) {
			err := ValidateOrderNumber(tt.number)
			if tt.wantErr {
				assert.Error(t, err, "ValidateOrderNumber(%s) = %v, want error: %v", tt.number, err, tt.wantErr)
			} else {
				assert.NoError(t, err, "ValidateOrderNumber(%s) = %v, want error: %v", tt.number, err, tt.wantErr)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		s    []string
		v    []string
		want bool
	}{
		{
			name: "return true when slice contains all values",
			s:    []string{"a", "b", "c"},
			v:    []string{"b"},
			want: true,
		},
		{
			name: "return true when slice contains all values with case-insensitive match",
			s:    []string{"a", "b", "c"},
			v:    []string{"B"},
			want: true,
		},
		{
			name: "return true when slice contains one of values",
			s:    []string{"a", "b", "c"},
			v:    []string{"a", "d"},
			want: true,
		},
		{
			name: "return true when slice contains one of values with case-insensitive match",
			s:    []string{"a", "b", "c"},
			v:    []string{"D", "C"},
			want: true,
		},
		{
			name: "return false when values is empty",
			s:    []string{"a", "b", "c"},
			v:    []string{},
			want: false,
		},
		{
			name: "return false when slice is empty",
			s:    []string{},
			v:    []string{"a"},
			want: false,
		},
		{
			name: "return false when slice does not contain one of values",
			s:    []string{"a", "b", "c"},
			v:    []string{"d"},
			want: false,
		},
		{
			name: "return false when slice does not contain one of values",
			s:    []string{"a", "b", "c"},
			v:    []string{"x", "y"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Contains(%v, %v): %s", tt.s, tt.v, tt.name), func(t *testing.T) {
			got := Contains(tt.s, tt.v...)
			assert.Equal(t, tt.want, got, "Contains(%v, %v) = %v, want %v", tt.s, tt.v, got, tt.want)
		})
	}
}

func TestFloat64Compare(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want int64
	}{
		{
			name: "return 1 when a is greater than b",
			a:    1.00001,
			b:    1.00000,
			want: 1,
		},
		{
			name: "return -1 when a is less than b",
			a:    1.00000,
			b:    1.00001,
			want: -1,
		},
		{
			name: "return 0 when a is equal to b",
			a:    1.00000,
			b:    1.00000,
			want: 0,
		},
		{
			name: "return 0 when a is greater than b but difference is within the precision threshold",
			a:    1.000001,
			b:    1.000000,
			want: 0,
		},
		{
			name: "return 0 when a is less than b but difference is within the precision threshold",
			a:    1.000000,
			b:    1.000001,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Float64Compare(%v, %v): %s", tt.a, tt.b, tt.name), func(t *testing.T) {
			got := Float64Compare(tt.a, tt.b)
			assert.Equal(t, tt.want, got, "Float64Compare(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		})
	}
}
