# How to run
```
docker-compose up --build
```
# How to check
```
curl -X POST localhost:8082/v1/compile -H "Content-Type: multipart/form-data" -F "upload[]=@./tests/main.tex" -F "token=asdf" > asdf.pdf
```
