# snowflake

A lock-free implementation of snowflake algorithm in Golang

## Get

`go get -u go.guoyk.net/snowflake`

## Usage

```go
// First argument should be a pre-defined zero time
// Second argument should be a unique unsigned integer with maximum 10 bits
s := snowflake.New(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), 0)

// Get a snowflake id
s.NewID()

// Stop and release all related resource
s.Stop()
```
 
## Credits
 
Guo Y.K., MIT License
