package cloneutil_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/cloneutil"
)

type cloneChild struct {
	Value string
}

type cloneSample struct {
	Scores []int
	Labels map[string][]string
	Child  *cloneChild
}

func TestClone_DeepCopiesCompositeValues(t *testing.T) {
	t.Parallel()

	original := cloneSample{
		Scores: []int{1, 2, 3},
		Labels: map[string][]string{"tags": {"alpha", "beta"}},
		Child:  &cloneChild{Value: "original"},
	}

	cloned, ok := cloneutil.Clone(original).(cloneSample)
	require.True(t, ok)

	want := cloneSample{
		Scores: []int{1, 2, 3},
		Labels: map[string][]string{"tags": {"alpha", "beta"}},
		Child:  &cloneChild{Value: "original"},
	}
	if diff := cmp.Diff(want, cloned); diff != "" {
		t.Errorf("Clone() mismatch (-want +got):\n%s", diff)
	}

	original.Scores[0] = 99
	original.Labels["tags"][0] = "changed"
	original.Child.Value = "changed"

	if diff := cmp.Diff(want, cloned); diff != "" {
		t.Errorf("Clone() shares mutable state (-want +got):\n%s", diff)
	}
}

func TestClone_PreservesSpecialValueSemantics(t *testing.T) {
	t.Parallel()

	instant := time.Date(2026, 5, 4, 12, 30, 0, 0, time.UTC)
	intValue := big.NewInt(42)
	floatValue := big.NewFloat(12.5)
	ratValue := big.NewRat(3, 7)

	clonedTime, ok := cloneutil.Clone(&instant).(*time.Time)
	require.True(t, ok)
	require.NotNil(t, clonedTime)
	assert.NotSame(t, &instant, clonedTime)
	assert.True(t, instant.Equal(*clonedTime))

	clonedInt, ok := cloneutil.Clone(intValue).(*big.Int)
	require.True(t, ok)
	require.NotNil(t, clonedInt)
	assert.NotSame(t, intValue, clonedInt)
	assert.Zero(t, intValue.Cmp(clonedInt))

	clonedFloat, ok := cloneutil.Clone(floatValue).(*big.Float)
	require.True(t, ok)
	require.NotNil(t, clonedFloat)
	assert.NotSame(t, floatValue, clonedFloat)
	assert.Zero(t, floatValue.Cmp(clonedFloat))

	clonedRat, ok := cloneutil.Clone(ratValue).(*big.Rat)
	require.True(t, ok)
	require.NotNil(t, clonedRat)
	assert.NotSame(t, ratValue, clonedRat)
	assert.Zero(t, ratValue.Cmp(clonedRat))

	intValue.SetInt64(99)
	floatValue.SetFloat64(99.5)
	ratValue.SetInt64(99)

	assert.Equal(t, int64(42), clonedInt.Int64())
	clonedFloatValue, _ := clonedFloat.Float64()
	assert.Equal(t, 12.5, clonedFloatValue)
	assert.Equal(t, "3/7", clonedRat.String())
}

func TestClone_NilValues(t *testing.T) {
	t.Parallel()

	assert.Nil(t, cloneutil.Clone(nil))

	var timePtr *time.Time
	clonedTime, ok := cloneutil.Clone(timePtr).(*time.Time)
	require.True(t, ok)
	assert.Nil(t, clonedTime)

	var intPtr *big.Int
	clonedInt, ok := cloneutil.Clone(intPtr).(*big.Int)
	require.True(t, ok)
	assert.Nil(t, clonedInt)
}
