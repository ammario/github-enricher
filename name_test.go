package main

import "testing"

func Test_cleanName(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nothing", args{"david"}, "david"},
		{"no weird symbols", args{"d@vid"}, "dvid"},
		{"no emails", args{"gel@microsoft.com"}, "gelmicrosoft.com"},
		{"Oliver", args{"Oliver"}, "Oliver"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanName(tt.args.n); got != tt.want {
				t.Errorf("cleanName() = %v, want %v", got, tt.want)
			}
		})
	}
}
