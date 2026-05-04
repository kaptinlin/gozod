package cloneutil_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/cloneutil"
)

func TestClone_PreservesNilMathPointers(t *testing.T) {
	t.Parallel()

	var floatPtr *big.Float
	clonedFloat, ok := cloneutil.Clone(floatPtr).(*big.Float)
	require.True(t, ok)
	assert.Nil(t, clonedFloat)

	var ratPtr *big.Rat
	clonedRat, ok := cloneutil.Clone(ratPtr).(*big.Rat)
	require.True(t, ok)
	assert.Nil(t, clonedRat)
}
