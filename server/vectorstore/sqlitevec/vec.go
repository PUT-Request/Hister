// SPDX-License-Identifier: AGPL-3.0-or-later

// Package sqlitevec bundles the sqlite-vec C extension and provides Auto() to
// register it with every new SQLite connection in the process.
//
// sqlite-vec source: https://github.com/asg017/sqlite-vec (v0.1.6)
// sqlite3.h source:  mattn/go-sqlite3 v1.14.42 (sqlite 3.51.3)
//
// To update bundled files run: update.sh
//
// The -Du_int*_t macros below are required for musl libc (Alpine Linux).
// musl does not define the POSIX u_int*_t aliases that sqlite-vec.c uses
// to typedef the C99 uint*_t types. The macros make those typedefs no-ops
// on musl and are harmless on glibc where the types are already compatible.
package sqlitevec

// #cgo CFLAGS: -DSQLITE_CORE
// #cgo CFLAGS: -Du_int8_t=uint8_t -Du_int16_t=uint16_t -Du_int64_t=uint64_t
// #cgo linux LDFLAGS: -lm
// #include "sqlite-vec.h"
import "C"

// Auto registers the sqlite-vec extension so that every future SQLite
// connection opened in this process has vec0 available.
func Auto() {
	C.sqlite3_auto_extension((*[0]byte)(C.sqlite3_vec_init))
}
