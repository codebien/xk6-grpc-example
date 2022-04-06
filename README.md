# xk6-grpc-example
Demo extension showing how to use the `lib/netext/grpcext` library directly from a k6 extension.

Build the binary using the `grpcext` branch:
```
xk6 build grpcext --with github.com/grafana/xk6-grpc-example=.
```

Run the built binary using the script included in the root:

```
./k6 run script.js
```

