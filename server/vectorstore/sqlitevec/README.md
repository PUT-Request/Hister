#  Local bundled wrapper for sqlite

## WARNING: This is required to be able to build sqlite-vec cross platform without dependencies

This package copies sqlite-vec.c and sqlite-vec.h into the repo and adds a one-liner sqlite3.h shim.
Then in the CGO preamble point at mattn's package dir via CGO_CFLAGS. This makes everything self-contained, no system dependency.

The downside is we're copying upstream C files and need to update them manually when sqlite-vec releases.
