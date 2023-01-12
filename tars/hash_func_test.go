package tars

import "testing"

func TestHashNew(t *testing.T) {
	testCases := []struct {
		name     string
		roomId   string
		wantHash uint32
	}{
		{
			name:     "#12723353",
			roomId:   "#12723353",
			wantHash: 1476478819,
		},
		{
			name:     "#12723353_native",
			roomId:   "#12723353_native",
			wantHash: 268793112,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if code := HashNew(tt.roomId); code != tt.wantHash {
				t.Errorf("HashNew() = %v, want %v", code, tt.wantHash)
			}
		})
	}
}
