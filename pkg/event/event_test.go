package event

import (
	"testing"
)

func TestGetEventInfo(t *testing.T) {
	type args struct {
		dns    string
		uri    string
		secret string
	}
	tests := []struct {
		name    string
		args    args
		want    *Info
		wantErr bool
	}{
		{
			name: "test happy path",
			args: args{
				dns:    "https://public-ip",
				uri:    "registry:dockerhub:codefresh:fortune:push",
				secret: "123456789",
			},
			want: &Info{
				Description: "Docker Hub codefresh/fortune push event",
				Endpoint:    "https://public-ip/nomios/dockerhub?secret=123456789",
				Status:      "active",
			},
			wantErr: false,
		},
		{
			name: "test complex path",
			args: args{
				dns:    "http://dev-tunneler.codefresh.io?client=itai-gendler&service=cfrouter",
				uri:    "registry:dockerhub:codefresh:fortune:push",
				secret: "123456789",
			},
			want: &Info{
				Description: "Docker Hub codefresh/fortune push event",
				Endpoint:    "http://dev-tunneler.codefresh.io//nomios/dockerhub?client=itai-gendler&secret=123456789&service=cfrouter",
				Status:      "active",
			},
			wantErr: false,
		},
		{
			name: "test bad dns",
			args: args{
				dns:    "::::",
				uri:    "registry:dockerhub:codefresh:fortune:push",
				secret: "123456789",
			},
			want: &Info{
				Description: "Docker Hub codefresh/fortune push event",
				Endpoint:    "",
				Status:      "active",
			},
			wantErr: false,
		},
		{
			name: "test bad event uri",
			args: args{
				dns:    "https://public-ip",
				uri:    "registry:dockerhub:unexpected-format:push",
				secret: "123456789",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEventInfo(tt.args.dns, tt.args.uri, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil &&
				got.Description != tt.want.Description &&
				got.Endpoint != tt.want.Endpoint {
				t.Errorf("GetEventInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
