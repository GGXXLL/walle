package _go

import "testing"

func Test_isRegularFile(t *testing.T) {
	tests := []struct {
		name    string
		f       string
		wantErr bool
	}{
		{"default", "helper.go", false},
		{"err", "helper123.go", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := isRegularFile(tt.f); (err != nil) != tt.wantErr {
				t.Errorf("isRegularFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
