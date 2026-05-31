package checks

import (
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestFileSizeChecks_ValidateHeadersAndAnnotate(t *testing.T) {
	t.Parallel()

	small := multipart.FileHeader{Size: 4}
	exact := multipart.FileHeader{Size: 8}
	large := multipart.FileHeader{Size: 12}

	tests := []struct {
		name      string
		check     core.ZodCheck
		valid     any
		invalid   any
		wantCode  core.IssueCode
		wantKey   string
		wantValue any
	}{
		{name: "minimum file size", check: MinFileSize(8), valid: &exact, invalid: &small, wantCode: core.TooSmall, wantKey: "minSize", wantValue: int64(8)},
		{name: "maximum file size", check: MaxFileSize(8), valid: exact, invalid: large, wantCode: core.TooBig, wantKey: "maxSize", wantValue: int64(8)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			issue := requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)
			assert.Equal(t, "file", issue.Origin())
			assert.False(t, issue.Inclusive())

			schema := newCheckAttachSchema(core.ZodTypeFile)
			bag := applyCheckAttach(t, tt.check, schema)
			assert.Equal(t, tt.wantValue, bag[tt.wantKey])
		})
	}
}

func TestFileSize_ReportsTooSmallAndTooBig(t *testing.T) {
	t.Parallel()

	check := FileSize(8)
	requireCheckAccepts(t, check, multipart.FileHeader{Size: 8})

	tooSmall := requireCheckRejects(t, check, multipart.FileHeader{Size: 4}, core.TooSmall)
	assert.Equal(t, int64(8), tooSmall.Minimum())
	assert.True(t, tooSmall.Inclusive())

	tooBig := requireCheckRejects(t, check, &multipart.FileHeader{Size: 12}, core.TooBig)
	assert.Equal(t, int64(8), tooBig.Maximum())
	assert.True(t, tooBig.Inclusive())

	schema := newCheckAttachSchema(core.ZodTypeFile)
	bag := applyCheckAttach(t, check, schema)
	assert.Equal(t, int64(8), bag["minSize"])
	assert.Equal(t, int64(8), bag["maxSize"])
	assert.Equal(t, int64(8), bag["size"])
}

func TestFileSizeChecks_ReadOSFileSize(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "upload.bin")
	require.NoError(t, os.WriteFile(path, []byte("gozod"), 0o600))

	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, file.Close())
	}()

	requireCheckAccepts(t, MinFileSize(5), file)
	requireCheckAccepts(t, MaxFileSize(5), *file)
}

func TestMime_ValidatesFileHeaderContentType(t *testing.T) {
	t.Parallel()

	allowed := []string{"image/png", "image/jpeg"}
	check := Mime(allowed)
	valid := multipart.FileHeader{Header: textproto.MIMEHeader{"Content-Type": []string{"image/png"}}}
	invalid := multipart.FileHeader{Header: textproto.MIMEHeader{"Content-Type": []string{"text/plain"}}}

	requireCheckAccepts(t, check, &valid)
	issue := requireCheckRejects(t, check, invalid, core.InvalidValue)
	assert.Equal(t, allowed[0], issue.Values()[0])
	assert.Equal(t, allowed[1], issue.Values()[1])

	notAFile := requireCheckRejects(t, check, "not a file", core.InvalidValue)
	assert.Equal(t, "not a file", notAFile.Input)

	schema := newCheckAttachSchema(core.ZodTypeFile)
	bag := applyCheckAttach(t, check, schema)
	got, ok := bag["mime"].([]string)
	require.True(t, ok)
	if diff := cmp.Diff(allowed, got); diff != "" {
		t.Errorf("Mime() bag mismatch (-want +got):\n%s", diff)
	}
}
