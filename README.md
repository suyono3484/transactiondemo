# Transaction Demo

This is a demo application server written in Go. It exposes two endpoints, `/add` and `/get`.
The `/add` endpoint is for adding new transaction while the `/get` endpoint retrieves transaction.

## Building
```shell
go build -o demo github.com/suyono3484/transactiondemo/cmd
```

## Running
```shell
./demo
```

Then the application will give an output something like this
```shell
HTTP server is listening on [::]:36707
```

### Hitting the APIs
Adding transaction:
```shell
curl -v -X POST -d "description=transaction%201&date=2023-09-12&amount=23.45" http://localhost:36707/add
```

Getting transaction:
```shell
curl -v "http://localhost:36707/get/5aa1031356d532b?target=Canada-Dollar"
```
output: `{"id":"5aa1031356d532b","description":"transaction 1","date":"2023-09-12","amount":23.45,"rate":1.326,"converted":31.09}`

## Testing
This project uses [Ginkgo v2](https://github.com/onsi/ginkgo). To run the Ginkgo test suite
```shell
go run github.com/onsi/ginkgo/v2/ginkgo
```