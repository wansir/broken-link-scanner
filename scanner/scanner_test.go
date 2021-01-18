package scanner

import "testing"

func Test_absURL(t *testing.T) {
	type args struct {
		currURL string
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				currURL: "../../installing-on-linux/introduction/multioverview/",
				baseURL: "https://kubesphere.com.cn/docs/pluggable-components/network-policy/",
			},
			want: "https://kubesphere.com.cn/docs/installing-on-linux/introduction/multioverview/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := absURL(tt.args.currURL, tt.args.baseURL); got != tt.want {
				t.Errorf("absURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
