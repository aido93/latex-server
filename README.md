# How to run
docker-compose.yml:
```
version: "2.0"
volumes:
  uploads:
services:
  latex:
    image: aido93/latex-server:latest
    container_name: latex-server
    ports:
      - 8082:8080
    volumes:
      - uploads:/data
```

After that just
```
docker-compose up -d
```
# How to check
```
curl -X POST localhost:8082/v1/compile -H "Content-Type: multipart/form-data" -F "upload[]=@./tests/main.tex" -F "token=asdf" > asdf.pdf
```
