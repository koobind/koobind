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

// A list of string which contains only unique items

type StringSet interface {
	DeepCopy() StringSet
	Add(string) StringSet
	AsList() []string
	Has(string) bool
	Len() int
}

type nothing struct{}

type stringSet struct {
	hash map[string]nothing
}

func NewStringSet() StringSet {
	return &stringSet{
		hash: make(map[string]nothing),
	}
}

func (this *stringSet) Has(s string) bool {
	_, exists := this.hash[s]
	return exists
}

func (this *stringSet) DeepCopy() StringSet {
	h := make(map[string]nothing)
	for k, v := range this.hash {
		h[k] = v
	}
	return &stringSet{
		hash: h,
	}
}

func (this *stringSet) Add(s string) StringSet {
	this.hash[s] = nothing{}
	return this
}

func (this *stringSet) AsList() []string {
	l := make([]string, 0, len(this.hash))
	for k, _ := range this.hash {
		l = append(l, k)
	}
	return l
}

func (this *stringSet) Len() int {
	return len(this.hash)
}
