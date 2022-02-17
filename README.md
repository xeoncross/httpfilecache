## HTTP File Cache (Go)

A simple HTTP request file cache that stores results by path in 
`$HOME/.cache` (or `$XDG_CACHE_HOME` if set). This is mostly for inspection 
use-cases like integrations tests or prototyping where you want to be able to 
easily browse a hierarchy of request bodies.

```
http://www.sub1.example.com/cool/webpage.html 

would be stored under:

~./cache
    /httpfilecache
        /.com
            /example.sub1.www
                /cool-webpage.html
```

The cache is shared globally between all projects using this library.

For production work loads consider https://github.com/gregjones/httpcache/ which
also allows the caching of byte ranges.