package ids

import (
	"reflect"
	"testing"

	"github.com/satori/go.uuid"
)

func TestToBase64(t *testing.T) {
	type args struct {
		id uuid.UUID
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"validNonNil", args{uuid.FromStringOrNil("ce629437-cce5-489f-856d-d44a291f72b3")}, "zmKUN8zlSJ-FbdRKKR9ysw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToBase64(tt.args.id); got != tt.want {
				t.Errorf("ToBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBase64ToUUID(t *testing.T) {
	type args struct {
		b64 string
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
	}{
		{"validNonNil", args{"zmKUN8zlSJ-FbdRKKR9ysw"}, uuid.FromStringOrNil("ce629437-cce5-489f-856d-d44a291f72b3"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base64ToUUID(tt.args.b64)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64ToUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Base64ToUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
	}{
		{"uuid", args{"ce629437-cce5-489f-856d-d44a291f72b3"}, uuid.FromStringOrNil("ce629437-cce5-489f-856d-d44a291f72b3"), false},
		{"b64", args{"zmKUN8zlSJ-FbdRKKR9ysw"}, uuid.FromStringOrNil("ce629437-cce5-489f-856d-d44a291f72b3"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseID(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseID() = %v, want %v", got, tt.want)
			}
		})
	}
}
