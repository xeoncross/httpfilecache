## HTTP File Cache (Go)

A simple HTTP request file cache that stores the full HTTP request payload on 
disk and returns it instead of making a request while the expires TTL is still
valid. Developed as a transparent, simple way to cache remote payloads without 
needing a database or caching logic.

It is only suitable for use as a 'private' cache (i.e. for a web-browser or an 
API-client and not for a shared proxy).

```
go get github.com/Xeoncross/httpfilecache
```

```
dir := "/whatever/you/want"
client := NewClient(dir, time.Second*60)

resp, err := client.Get("https://example.com/api/here?id=1234") // cached on first request
resp, err := client.Get("https://example.com/api/here?id=1234") // loaded from disk after that
```

Also consider https://github.com/gregjones/httpcache/ which supports a different
feature set.