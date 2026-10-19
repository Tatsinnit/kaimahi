//go:build windows

package materialio

import "syscall"

const readFlags = syscall.FILE_FLAG_OPEN_REPARSE_POINT
