# goshort [![builds.sr.ht status](https://builds.sr.ht/~delthas/goshort.svg)](https://builds.sr.ht/~delthas/goshort?)

A trivially small link shortener.

Usage: `goshort -port 8080 -url "https://l.dille.cc"`

Shorten a link by visiting: `<server_url>/<short>/<url>`, the shortened URL will be `<server_url>/<short>`.

As a special case `<server_url>/hash/<url>` generates a small key for you (hex).

Pre-built binaries available at:

| OS | URL  |
|---|---|
| Linux x64 | https://delthas.fr/goshort/linux/goshort  |
| Mac OS X  x64 |  https://delthas.fr/goshort/mac/goshort |
| Windows  x64 |  https://delthas.fr/goshort/windows/goshort.exe |
