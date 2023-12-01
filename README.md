# syntryve

## Requirements

### Go 
`go1.20` or higher
```
go version
> go version go1.21.3 darwin/arm64
```

Build `syntryve`
```
make
```

_(Optional)_ Put it into your machine's PATH
```
cp build/syntryve ~/go/bin/syntryve
```

## Usage 
```
syntryve serve --nats-url="NATS_URL" --stream-url="STREAM_URL" --token="ACCESS_TOKEN"
```

## Example
```
syntryve serve --nats-url="nats://34.107.87.29" --stream-url="archway-sandbox.archway.block" --token="SAFTESTASRRQKYVET7P56VJT76AKSYNTROPY"
```
