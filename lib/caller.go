package ecslogs

import (
	"path/filepath"
	"runtime"
	"strings"
)

func caller(depth int) (file string, line int, function string, ok bool) {
	var fn *runtime.Func
	var pc uintptr
	var pkg string

	if pc, file, line, ok = runtime.Caller(depth + 1); !ok {
		return
	}

	file = filepath.Base(file)

	if fn = runtime.FuncForPC(pc); fn != nil {
		pkg, function = parseFunctionName(fn.Name())
		file = filepath.Join(pkg, file)
	}

	return
}

func parseFunctionName(name string) (pkg string, fn string) {
	var pkgOff int
	var fnOff int

	if off := strings.LastIndexByte(name, '('); off >= 0 {
		// This is a method name because has its target type defined
		// between parenthesis.
		if pkgOff, fnOff = off, off; pkgOff > 0 && name[pkgOff-1] == '.' {
			pkgOff--
		}
	} else if off = strings.LastIndexByte(name, '.'); off >= 0 {
		// This is a simple function name, the package and function
		// names are separated by a dot
		pkgOff, fnOff = off, off+1
	} else {
		// We couldn't figure out the format, simply return the full
		// string as the function name and leave the package blank.
		fn = name
		return
	}

	pkg, fn = name[:pkgOff], name[fnOff:]
	return
}
