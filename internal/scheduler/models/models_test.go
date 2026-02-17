package models

import "testing"

func TestTableNames(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"user", (User{}).TableName(), "users"},
		{"driver", (Driver{}).TableName(), "drivers"},
		{"request", (Request{}).TableName(), "requests"},
		{"shift", (Shift{}).TableName(), "shifts"},
		{"shift_request", (ShiftRequest{}).TableName(), "shift_requests"},
		{"shift_staff", (ShiftStaff{}).TableName(), "shift_staffs"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Fatalf("%s: got %s want %s", tc.name, tc.got, tc.want)
		}
	}
}
