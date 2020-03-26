# Automony API Server


## Build

Simply use `go build` to build the binary

## Migrate DB Schema

First, you need to prepare a postgres server.

Configurate your database by setting environment variable:

```
$ export AUTONOMY_ORM_CONN='user=user password= host=localhost port=5432 dbname=autonomy connect_timeout=10 sslmode=disable'
```

Step into folder `schema/command/migrate` and run

```
$ go run main.go
```

The database should be configured well.

## Generate JWT private key

Use ssh-keygen to generate an RSA key with a passphrase:

```
$ ssh-keygen -t rsa -b 4096 -m PEM -f jwt-sign.key
```

## Prepare a bitmark account V2 seed

Use bitmark sdk to generate a seed for the server

## Run

Before you run the app you need to set the following environment variables:

- Assign a bitmark seed to the server:

```
$ export AUTONOMY_ACCOUNT_SEED='9J874PsxvHV7tSG69bwS75gBVoeRWPhhM' // ONLY USE FOR TEST
```

- Set the key path and the key phrase:

```
$ export AUTONOMY_JWT_KEYFILE='./jwt-sign.key'
$ export AUTONOMY_JWT_PASSWORD='the-sign-key-passphrase'
```

- Configurate the postgres db

```
$ export AUTONOMY_ORM_CONN='user=user password= host=localhost port=5432 dbname=autonomy connect_timeout=10 sslmode=disable'
```

Finally, you need to make a copy of the file `config.sample.yaml` to `config.yaml` with proper configurate.

Now, You can run the server:

```
./autonomy-api
```
