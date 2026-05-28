package logger_test

import (
	"testing"

	"github.com/amfib87/go_musthave-GophKeeper/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "success initialization",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.Initialize()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, logger.Tlog{}, log)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, log.Log)
			}
		})
	}
}
