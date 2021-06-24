# go-tool
A toolkit for go coding

- api : go api project building tool
- code : go coding tool
- logfmt: api log formatter
- ws: websocket protocol tool

## api tool
### command
```shell
api --help # see help
api compile --help # see compile help
api swagger --help # see generate doc help
```

### api doc json tags
Add api doc tags in struct field tags for specified functions, use `,`to split keys in same tag.
example:

```go
type Sample struct {
	BaseParams `api_doc:"skip"`
	Time time.Time `json:"time" api_doc:"type=string"`
}
```



| key  | desc                                                         | remark |
| ---- | ------------------------------------------------------------ | ------ |
| skip | Field will not generate desc in swagger                      |        |
| type | Specify swagger doc display type for purpose. In example above, `Sample.Time`'s type will be parsed as string, not `struct{}` in swagger doc |        |
|      |                                                              |        |

