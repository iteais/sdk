package utils

import "testing"

func TestCheckIpsInSameSubnet(t *testing.T) {
	type args struct {
		clientIP string
		serverIP string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ips_in_same_subnet_false",
			args: args{
				clientIP: "89.0.142.86",
				serverIP: "244.178.44.111",
			},
			want: false,
		},
		{
			name: "ips_in_same_subnet_true",
			args: args{
				clientIP: "89.0.142.86",
				serverIP: "89.0.142.87",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckIpsInSameSubnet(tt.args.clientIP, tt.args.serverIP); got != tt.want {
				t.Errorf("CheckIpsInSameSubnet() = %v, want %v", got, tt.want)
			}
		})
	}
}
