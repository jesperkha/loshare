# loshare

File sharing over local networks.

## Use

This project is meant to run on a raspberry pi or similar home server.

```sh
# Create .env and fill in PORT
cenv fix

# Run
go run cmd/main.go
```

A `.dump` folder will be create and used to temporarily store files.
