package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileOutput(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  OutputConfig
		wantErr bool
	}{
		{
			name: "valid file path",
			config: OutputConfig{
				Type:       OutputTypeFile,
				FilePath:   filepath.Join(tmpDir, "audit.log"),
				MaxSize:    1024 * 1024,
				MaxBackups: 5,
				MaxAge:     7,
			},
			wantErr: false,
		},
		{
			name: "empty file path",
			config: OutputConfig{
				Type:     OutputTypeFile,
				FilePath: "",
			},
			wantErr: true,
		},
		{
			name: "path traversal attempt",
			config: OutputConfig{
				Type:     OutputTypeFile,
				FilePath: "../../../etc/passwd",
			},
			wantErr: true,
		},
		{
			name: "valid path with defaults",
			config: OutputConfig{
				Type:     OutputTypeFile,
				FilePath: filepath.Join(tmpDir, "test.log"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := newFileOutput(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)

			// Clean up
			if output != nil {
				output.Close()
			}
		})
	}
}

func TestFileOutput_Write(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "audit.log")

	output, err := newFileOutput(OutputConfig{
		Type:       OutputTypeFile,
		FilePath:   logFile,
		MaxSize:    1024,
		MaxBackups: 3,
		MaxAge:     7,
	})
	require.NoError(t, err)
	defer output.Close()

	event := AuditEvent{
		Timestamp:    time.Now(),
		EventType:    EventTypeToolExecution,
		Category:     CategoryAccess,
		Outcome:      OutcomeSuccess,
		Username:     "test-user",
		Action:       "test-action",
		ResourceType: "test-resource",
	}

	err = output.Write(event)
	require.NoError(t, err)

	// Verify file was written
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestFileOutput_CompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.log")
	dstFile := filepath.Join(tmpDir, "dest.log.gz")

	// Create source file
	err := os.WriteFile(srcFile, []byte("test log content"), 0644)
	require.NoError(t, err)

	output := &fileOutput{
		config: OutputConfig{
			FilePath: filepath.Join(tmpDir, "audit.log"),
		},
	}

	// Test successful compression
	err = output.compressFile(srcFile, dstFile)
	require.NoError(t, err)

	// Verify compressed file exists
	_, err = os.Stat(dstFile)
	require.NoError(t, err)
}

func TestFileOutput_CompressFile_PathValidation(t *testing.T) {
	output := &fileOutput{}

	tests := []struct {
		name    string
		src     string
		dst     string
		wantErr bool
	}{
		{
			name:    "empty source path",
			src:     "",
			dst:     "/tmp/dest.gz",
			wantErr: true,
		},
		{
			name:    "empty destination path",
			src:     "/tmp/src.log",
			dst:     "",
			wantErr: true,
		},
		{
			name:    "path traversal in source",
			src:     "../../../etc/passwd",
			dst:     "/tmp/dest.gz",
			wantErr: true,
		},
		{
			name:    "path traversal in destination",
			src:     "/tmp/src.log",
			dst:     "../../../etc/passwd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := output.compressFile(tt.src, tt.dst)
			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}
