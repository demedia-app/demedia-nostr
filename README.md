# DeMedia Poc

## How to run

You have to run a hub and peers on two or more terminals

### Hub

Open a terminal, then

```shell
cd hub
```
```shell
go run main.go
```

### Peer

Open a terminal, then

```shell
cd peer
```
```shell
go run main.go
```

### Endpoints

Get peers mapping
```shell
curl --location --request GET '0.0.0.0:8080/peer'
```

Get all todos of particular peer
```shell
curl --location --request GET '0.0.0.0:8080/todo' \
--header 'Peer: QmcQH5NGNk665u33ZP9zuzYVor6nbF2aopFeCkbsaBSsCs'
```

Create a todo on particular peer
```shell
curl --location --request POST '0.0.0.0:8080/todo' \
--header 'Peer: QmeHyXp4n72gUgCK6je9jcHVzNV9wAxACVPWRrUiwWswVJ' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": "5d48c042-3ae6-4b9d-9c6f-37d58c8922a3",
    "task": "mamama ananan",
    "title": "7fhfhfh"
}'
```

Fetch data sending SQL query
```shell
curl --location --request POST '0.0.0.0:8080/fetch' \
--header 'Peer: 16Uiu2HAkwyAuL5KbJ9kGZ12ubf8PQ5oumVC2gxLRgasn9yw4cg6F' \
--header 'Content-Type: application/json' \
--data-raw '{
   "query": "SELECT * FROM \"16Uiu2HAmP44YB5WWWdYccDYRzByum6fWDma13csdVUcySzwPMqYx_todos\" WHERE id = '\''6337c26a-f59d-4830-840f-91ff4918bc35'\''"
}'
```

Upload file to blob
```shell
curl --location --request POST '0.0.0.0:8080/file' \
--form 'file=@"/Users/sithumsandeepa/Downloads/1659222142514-Revel Food Corner_Table K-4_2022-07-30.png"'
```

### MinIO Docker

```dockerfile
mkdir -p ~/minio/data &&
docker run \
   -p 9000:9000 \
   -p 9090:9090 \
   --name minio \
   -v ~/minio/data:/data \
   -e "MINIO_ROOT_USER=ROOTNAME" \
   -e "MINIO_ROOT_PASSWORD=CHANGEME123" \
   quay.io/minio/minio server /data --console-address ":9090"
```

Set env like below

```shell
export AWS_ACCESS_KEY_ID=accessKeyID
export AWS_SECRET_ACCESS_KEY=secretAccessKey
export BUCKET_URI="s3://codepipeline-ap-south-1-61245200273?region=ap-south-1"
export POSTGRESQL_DATABASE="postgres://nostr:nostr@0.0.0.0:5432/nostr?sslmode=disable"
```
