# Alerty Go

`alerty-go` is a Go package designed to help developers automatically capture and report errors. It is suitable for both backend and frontend Go applications, making it easier to monitor and debug applications in various environments.

## Features

- Error capturing
- Automatic crash or panic capturing
- Easy to use API

## Installation

To install `alerty-go`, run:

```bash
go get github.com/alerty-ai/alerty-go
```

## Usage

Import the package and use it in your application to capture and report errors:

```go
import (
    "errors"
    "github.com/alerty-ai/alerty-go/pkg/alerty"
)

func main() {
    // Initialize the tracer
    alerty.Start(alerty.AlertyServiceConfig{
        Name:        "your-service-name",
        Version:     "1.0.0",
        Environment: "production",
    })
    defer alerty.Stop()

    // Capture an error
    err := doSomething()
    if err != nil {
        alerty.CaptureError(err)
    }
}

func doSomething() error {
    // Simulate an error
    return errors.New("something went wrong")
}
```

## Initialization

You need to initialize the tracer with the configuration for your Alerty service.

```go
alerty.Start(alerty.AlertyServiceConfig{
    Name:        "your-service-name",
    Version:     "1.0.0",
    Environment: "production",
})
```

## Error Capturing

To capture an error, use the `CaptureError` function:

```go
alerty.CaptureError(err)
```

This will create a span, record the exception, and end the span, effectively reporting the error to your Alerty backend.

## Panic Recovery

To capture panics, use the `CapturePanic` function inside a `defer` block:

```go
defer func() {
    if r := recover(); r != nil {
        alerty.CapturePanic(r)
    }
}()

panic("an example panic")
```

Or use the `Recover` function to handle panic recovery:

```go
defer alerty.Recover(func(r interface{}) {
    fmt.Println(r)
})

panic("an example panic")
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
