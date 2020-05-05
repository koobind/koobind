package utils

import (
	"sort"
	"testing"
)
import "github.com/stretchr/testify/assert"

func Test1(t *testing.T) {
	x := NewStringSet().Add("s1").Add("s2").Add("s1")
	assert.Equal(t, 2, x.Len())
	assert.True(t, x.Has("s1"))
	assert.True(t, x.Has("s2"))
	assert.False(t, x.Has("s3"))
	l := x.AsList()
	assert.Equal(t, 2, len(l))
	sort.Strings(l)
	assert.Equal(t, "s1", l[0])
	assert.Equal(t, "s2", l[1])
	assert.Equal(t, 2, x.Len())
}

func Test2(t *testing.T) {
	x := NewStringSet().Add("s1").Add("s2").Add("s1")
	y := x.DeepCopy().Add("s3")
	assert.Equal(t, 2, x.Len())
	assert.Equal(t, 3, y.Len())

	assert.True(t, x.Has("s1"))
	assert.True(t, x.Has("s2"))
	assert.False(t, x.Has("s3"))

	assert.True(t, y.Has("s1"))
	assert.True(t, y.Has("s2"))
	assert.True(t, y.Has("s3"))
}
