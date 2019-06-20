package aster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangePkgName(t *testing.T) {
	cases := []struct {
		code     string
		pkgname  string
		expected string
	}{
		{"package abc\n", "xyz", "package xyz\n"},
		{"package a_b_c\n", "xyz", "package xyz\n"},
		{"package abc \r \n", "xyz", "package xyz \r \n"},
		{"\n\npackage abc \r \n", "xyz", "\n\npackage xyz \r \n"},
		{"\n\npackage    abc \r \n", "xyz", "\n\npackage xyz \r \n"},
		{"\n\npackage\t    abc \r \n", "xyz", "\n\npackage xyz \r \n"},
		{"package abc//\n", "xyz", "package xyz//\n"},
		{"package abc //\n", "xyz", "package xyz //\n"},
		{"package    abc // comment\n", "xyz", "package xyz // comment\n"},
	}
	for _, c := range cases {
		actual := ChangePkgName(c.code, c.pkgname)
		assert.Equal(t, c.expected, actual)
	}
}
