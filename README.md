# golang-infakt

Go client library for the [inFakt API](https://www.infakt.pl/) — a Polish invoicing and accounting service.

## Installation

```bash
go get github.com/przemekperon/golang-infakt
```

## Usage

```go
package main

import "github.com/przemekperon/golang-infakt"

func main() {
    client := infakt.NewClient("your-api-key")
    _ = client
}
```

## License

[MIT](LICENSE)
