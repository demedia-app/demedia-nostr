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

