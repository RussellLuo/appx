# appx

An application framework that makes it easy to build pluggable and reusable applications.


## Installation

```bash
$ go get -u github.com/RussellLuo/appx
```


## Application Lifecycle

Simple Application:

```
            +-------+                   +-------+
(BEGIN) --> | Init  | --> (WORKING) --> | Clean | --> (END)
            +-------+                   +-------+
```

Runnable Application:

```
            +-------+     +-------+                   +-------+     +-------+
(BEGIN) --> | Init  | --> | Start | --> (RUNNING) --> | Stop  | --> | Clean | --> (END)
            +-------+     +-------+                   +-------+     +-------+
```

## Examples

- [Basic usage](example_test.go)
- [HTTP applications][1]
- [HTTP applications][2] ([Gin][3]-based)
- [CRON applications][4]


## Documentation

Checkout the [Godoc][5].


## License

[MIT](LICENSE)


[1]: https://github.com/RussellLuo/kok/blob/master/pkg/appx/httpapp/example_test.go
[2]: https://gist.github.com/RussellLuo/5e706323e215bd8cb840cb7ae6aabae7
[3]: https://github.com/gin-gonic/gin
[4]: https://github.com/RussellLuo/kok/blob/master/pkg/appx/cronapp/example_test.go
[5]: https://pkg.go.dev/mod/github.com/RussellLuo/appx
