---
'@jakebailey/dprint-plugin-gofumpt': patch
---

Fix a crash on deeply nested code, like a long chain of string concatenations,
by using a 1 MiB stack instead of 64 KiB.
