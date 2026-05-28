package service_test

import (
	"testing"

	"github.com/amfib87/go_musthave-GophKeeper/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestGetIDfromPath(t *testing.T) {
	tests := []struct {
		name     string
		pathURL  string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple path with ID",
			pathURL:  "/api/users/123",
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "path with trailing slash",
			pathURL:  "/api/users/123/",
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "path with multiple trailing slashes",
			pathURL:  "/api/users/123///",
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "empty string",
			pathURL:  "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "single segment",
			pathURL:  "users",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "two segments",
			pathURL:  "/api/users",
			expected: "users",
			wantErr:  false,
		},
		{
			name:     "complex path with nested segments",
			pathURL:  "/v1/api/users/profile/settings/abc123xyz",
			expected: "abc123xyz",
			wantErr:  false,
		},
		{
			name:     "path with special characters in ID",
			pathURL:  "/api/items/item-123_abc",
			expected: "item-123_abc",
			wantErr:  false,
		},
		{
			name:     "IP address as path",
			pathURL:  "192.168.1.100",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetIDfromPath(tt.pathURL)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
