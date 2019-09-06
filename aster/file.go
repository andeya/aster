package aster

import (
	"go/ast"
)

// File the 'ast.File' with filename and fileSet.
type File struct {
	Filename string
	*ast.File
	*PackageInfo
}
