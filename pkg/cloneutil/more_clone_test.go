package cloneutil_test

import (
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/cloneutil"
)

func TestClone_CopiesInterfaceAndArrayValues(t *testing.T) {
	t.Parallel()

	var original any = []map[string]int{{"a": 1}, {"b": 2}}
	cloned, ok := cloneutil.Clone(original).([]map[string]int)
	require.True(t, ok)

	want := []map[string]int{{"a": 1}, {"b": 2}}
	if diff := cmp.Diff(want, cloned); diff != "" {
		t.Errorf("Clone() mismatch (-want +got):\n%s", diff)
	}

	original.([]map[string]int)[0]["a"] = 99
	if diff := cmp.Diff(want, cloned); diff != "" {
		t.Errorf("Clone() shares interface-backed state (-want +got):\n%s", diff)
	}

	array := [2][]int{{1, 2}, {3, 4}}
	clonedArray, ok := cloneutil.Clone(array).([2][]int)
	require.True(t, ok)
	array[0][0] = 99

	wantArray := [2][]int{{1, 2}, {3, 4}}
	if diff := cmp.Diff(wantArray, clonedArray); diff != "" {
		t.Errorf("Clone() array mismatch (-want +got):\n%s", diff)
	}
}

func TestClone_CopiesNonPointerSpecialValues(t *testing.T) {
	t.Parallel()

	intValue := *big.NewInt(42)
	floatValue := *big.NewFloat(12.5)
	ratValue := *big.NewRat(3, 7)

	clonedInt, ok := cloneutil.Clone(intValue).(big.Int)
	require.True(t, ok)
	assert.Zero(t, intValue.Cmp(&clonedInt))

	clonedFloat, ok := cloneutil.Clone(floatValue).(big.Float)
	require.True(t, ok)
	assert.Zero(t, floatValue.Cmp(&clonedFloat))

	clonedRat, ok := cloneutil.Clone(ratValue).(big.Rat)
	require.True(t, ok)
	assert.Zero(t, ratValue.Cmp(&clonedRat))
}

func TestClone_PreservesNilCompositeTypes(t *testing.T) {
	t.Parallel()

	var slice []int
	clonedSlice, ok := cloneutil.Clone(slice).([]int)
	require.True(t, ok)
	assert.Nil(t, clonedSlice)

	var items map[string]int
	clonedMap, ok := cloneutil.Clone(items).(map[string]int)
	require.True(t, ok)
	assert.Nil(t, clonedMap)

	var value any
	typedInterface := value
	assert.Nil(t, cloneutil.Clone(typedInterface))
}
