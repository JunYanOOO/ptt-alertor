package controllers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Ptt-Alertor/ptt-alertor/config"
	"github.com/Ptt-Alertor/ptt-alertor/models/top"
)

func TestTemplatesHideDisabledFeatures(t *testing.T) {
	features := config.Features{KeywordTracking: true}
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "line.html",
			data: struct {
				URI, WSHost, S3Domain string
				Count                 []string
				Features              config.Features
			}{Features: features},
		},
		{
			name: "docs.html",
			data: struct {
				URI, S3Domain string
				Features      config.Features
			}{Features: features},
		},
		{
			name: "top.html",
			data: struct {
				URI, S3Domain     string
				Keywords, Authors top.WordOrders
				PushSum           top.WordOrders
				Features          config.Features
			}{Features: features},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := appTemplates().ExecuteTemplate(&output, test.name, test.data); err != nil {
				t.Fatalf("execute template: %v", err)
			}
			for _, hidden := range []string{"新增作者", "新增推文數", "新增噓文數", "推文追蹤"} {
				if strings.Contains(output.String(), hidden) {
					t.Errorf("disabled feature text %q is visible", hidden)
				}
			}
		})
	}
}
