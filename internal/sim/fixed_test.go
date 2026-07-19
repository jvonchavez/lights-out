package sim

import "testing"

func TestMilliMul(t *testing.T) {
	tests := []struct {
		name       string
		a, b, want Milli
	}{
		{"identity", 1500, One, 1500},
		{"half times half", 500, 500, 250},
		{"zero", 0, 12345, 0},
		{"negative", -1500, 2000, -3000},
		{"truncates toward zero", 1, 1, 0},
		{"negative truncates toward zero", -1, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Mul(tt.b); got != tt.want {
				t.Errorf("%d.Mul(%d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMilliDiv(t *testing.T) {
	tests := []struct {
		name       string
		a, b, want Milli
	}{
		{"identity", 1500, One, 1500},
		{"half by two", 500, 2000, 250},
		{"negative", -3000, 2000, -1500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Div(tt.b); got != tt.want {
				t.Errorf("%d.Div(%d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestFromInt(t *testing.T) {
	if got := FromInt(50); got != 50000 {
		t.Errorf("FromInt(50) = %d, want 50000", got)
	}
}
