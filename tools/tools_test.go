package tools

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

func TestPkgName(t *testing.T) {
	cases := []struct {
		filenameOrDirectory string
		src                 interface{}
		expected            string
	}{
		{"", "package abc\n", "abc"},
		{"", "package a_b_c\n", "a_b_c"},
		{"", "package abc \r \n", "abc"},
		{"", "\n\npackage abc \r \n", "abc"},
		{"", "\n\npackage    abc \r \n", "abc"},
		{"", "\n\npackage\t    abc \r \n", "abc"},
		{"", "package abc//\n", "abc"},
		{"", "package abc //\n", "abc"},
		{"", "package    abc // comment\n", "abc"},
		{"util.go", nil, "aster"},
		{"./", nil, "aster"},
	}
	for _, c := range cases {
		actual, err := PkgName(c.filenameOrDirectory, c.src)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, actual)
	}
}

func TestFormat(t *testing.T) {
	const code = `
	package z
	import "fmt"
	var a=0
	`
	const expected = `package z

var a = 0
`

	b, err := Format("", code, nil)
	assert.NoError(t, err)
	actual := string(b)
	assert.Equal(t, expected, actual)
}
