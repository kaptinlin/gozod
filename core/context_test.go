package core

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePayload_AddIssuePrefixesCurrentPath(t *testing.T) {
	t.Parallel()

	path := []any{"user"}
	payload := NewParsePayloadWithPath("value", path)
	path[0] = "mutated"
	payload.PushPath("email")
	payload.AddIssue(ZodRawIssue{Code: InvalidType, Path: []any{"domain"}})

	issues := payload.Issues()
	require.Len(t, issues, 1)

	wantPath := []any{"user", "email", "domain"}
	if diff := cmp.Diff(wantPath, issues[0].Path); diff != "" {
		t.Errorf("AddIssue() path mismatch (-want +got):\n%s", diff)
	}
}

func TestParsePayload_AddIssueWithPathUsesOverrideCopy(t *testing.T) {
	t.Parallel()

	payload := NewParsePayload("value")
	override := []any{"items", 0}
	payload.AddIssueWithPath(ZodRawIssue{Code: Custom}, override)
	override[0] = "mutated"

	issues := payload.Issues()
	require.Len(t, issues, 1)

	wantPath := []any{"items", 0}
	if diff := cmp.Diff(wantPath, issues[0].Path); diff != "" {
		t.Errorf("AddIssueWithPath() path mismatch (-want +got):\n%s", diff)
	}
}

func TestParsePayload_CleanClonePreservesValueAndPathWithoutIssues(t *testing.T) {
	t.Parallel()

	payload := NewParsePayloadWithPath("value", []any{"field"})
	payload.AddIssueWithMessage("invalid")
	clean := payload.WithCleanIssues()

	assert.Equal(t, "value", clean.Value())
	assert.False(t, clean.HasIssues())
	assert.Equal(t, 1, payload.IssueCount())

	wantPath := []any{"field"}
	if diff := cmp.Diff(wantPath, clean.Path()); diff != "" {
		t.Errorf("WithCleanIssues() path mismatch (-want +got):\n%s", diff)
	}
}

func TestParsePayload_CloneCopiesSlices(t *testing.T) {
	t.Parallel()

	payload := NewParsePayloadWithPath("value", []any{"field"})
	payload.AddIssueWithCode(Custom, "invalid")
	cloned := payload.Clone()

	payload.PushPath("nested")
	payload.AddIssueWithMessage("second")

	assert.Equal(t, 1, cloned.IssueCount())
	wantPath := []any{"field"}
	if diff := cmp.Diff(wantPath, cloned.Path()); diff != "" {
		t.Errorf("Clone() path mismatch (-want +got):\n%s", diff)
	}
}

func TestParseContext_CopyModifiersPreserveExistingFields(t *testing.T) {
	t.Parallel()

	customError := func(ZodRawIssue) string { return "custom" }
	base := &ParseContext{ReportInput: true}

	withError := base.WithCustomError(customError)
	require.NotNil(t, withError.Error)
	assert.Equal(t, "custom", withError.Error(ZodRawIssue{}))
	assert.True(t, withError.ReportInput)
	assert.Nil(t, base.Error)

	withoutReportInput := withError.WithReportInput(false)
	require.NotNil(t, withoutReportInput.Error)
	assert.Equal(t, "custom", withoutReportInput.Error(ZodRawIssue{}))
	assert.False(t, withoutReportInput.ReportInput)
	assert.True(t, withError.ReportInput)

	cloned := withoutReportInput.Clone()
	require.NotSame(t, withoutReportInput, cloned)
	require.NotNil(t, cloned.Error)
	assert.Equal(t, "custom", cloned.Error(ZodRawIssue{}))
	assert.False(t, cloned.ReportInput)
}

func TestParsePayload_AddIssuesAppendsIssues(t *testing.T) {
	t.Parallel()

	payload := NewParsePayload("value")
	payload.AddIssues()
	assert.False(t, payload.HasIssues())

	payload.AddIssues(
		ZodRawIssue{Code: InvalidType, Message: "first"},
		ZodRawIssue{Code: Custom, Message: "second"},
	)

	issues := payload.Issues()
	require.Len(t, issues, 2)
	assert.Equal(t, InvalidType, issues[0].Code)
	assert.Equal(t, Custom, issues[1].Code)
}

func TestParsePayload_AccessorsDefensivelyCopySlices(t *testing.T) {
	t.Parallel()

	payload := NewParsePayloadWithPath("value", []any{"user"})
	payload.AddIssue(ZodRawIssue{Code: Custom, Path: []any{"name"}})

	path := payload.Path()
	path[0] = "mutated"
	issues := payload.Issues()
	issues[0].Code = InvalidType
	issues[0].Path[0] = "mutated"

	wantPath := []any{"user"}
	if diff := cmp.Diff(wantPath, payload.Path()); diff != "" {
		t.Errorf("Path() mismatch (-want +got):\n%s", diff)
	}

	current := payload.Issues()
	require.Len(t, current, 1)
	assert.Equal(t, Custom, current[0].Code)
	wantIssuePath := []any{"user", "name"}
	if diff := cmp.Diff(wantIssuePath, current[0].Path); diff != "" {
		t.Errorf("Issues() path mismatch (-want +got):\n%s", diff)
	}
}

func TestParsePayload_SettersAndContext(t *testing.T) {
	t.Parallel()

	payload := NewParsePayload("before")
	payload.SetValue("after")
	assert.Equal(t, "after", payload.Value())

	issues := []ZodRawIssue{{Code: Custom, Message: "custom"}}
	payload.SetIssues(issues)
	issues[0].Code = InvalidType

	current := payload.Issues()
	require.Len(t, current, 1)
	assert.Equal(t, Custom, current[0].Code)

	ctx := NewParseContext().WithReportInput(true)
	payload.SetContext(ctx)
	assert.Same(t, ctx, payload.Context())
}
