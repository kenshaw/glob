# D14. Path and Host are globs with their own options

Status: Proposed.

`notes.txt` of the maintainers asks for a type for paths and a type for host
names. The types embed `Glob`, so a struct field of either type unmarshals
from text with the right options.

1. `Path` has the separator `/` and the globstar option (D16). `Path.Match`
   changes each separator of the operating system to `/` first, with
   `filepath.ToSlash`.
2. `Host` has the separator `.` and the fold option, because a host name has
   no case. `Host.Match` and `CompileHost` ignore one dot at the end of a
   name, so `example.com.` matches as `example.com` does.
3. `CompilePath` and `CompileHost` take more options after the ones of the
   type.
