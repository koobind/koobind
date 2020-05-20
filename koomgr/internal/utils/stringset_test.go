/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

  koobind is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  koobind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/
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
