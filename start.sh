go build -o webshield src/main.go
export hostname=yourhostname.com
export LOGGING=DEBUG
export PORT=9865
export DOT_SERVER_DISABLED=true
./webshield
