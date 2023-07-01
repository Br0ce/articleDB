package ids

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestUniqueID(t *testing.T) {
	t.Parallel()

	got := UniqueID()

	if _, err := uuid.Parse(got); err != nil {
		t.Errorf("invalid ID, got %v, err %v", got, err.Error())
	}

	other := UniqueID()
	if got == other {
		t.Errorf("id is not unique, got %v, other %v", got, other)
	}

}

func TestValidID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "valid id",
			id:   uuid.NewString(),
			want: true,
		},
		{
			name: "empty id",
			id:   "",
			want: false,
		},
		{
			name: "invalid",
			id:   "1234",
			want: false,
		},
		{
			name: "invalid format",
			id:   "67a3f9158d5e-49d8-9ecb-03471dc7619f",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(uuid.NewString())
			if got := ValidID(tt.id); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
