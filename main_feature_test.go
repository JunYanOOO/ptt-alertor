package main

import (
	"net/http"
	"testing"

	"github.com/Ptt-Alertor/ptt-alertor/config"
)

func TestRegisterTrackingAPIsKeywordOnly(t *testing.T) {
	router := newRouter()
	registerTrackingAPIs(router, config.Features{KeywordTracking: true})

	tests := []struct {
		path string
		want bool
	}{
		{path: "/keyword/boards", want: true},
		{path: "/author/boards", want: false},
		{path: "/pushsum/boards", want: false},
		{path: "/articles", want: false},
	}
	for _, test := range tests {
		handle, _, _ := router.Lookup(http.MethodGet, test.path)
		if got := handle != nil; got != test.want {
			t.Errorf("route %s registered = %t, want %t", test.path, got, test.want)
		}
	}
}
