## Run on MacOS (Tested on M3)
you must extract the contents of: https://github.com/AirspaceTechnologies/or-tools/releases/download/v9.10-go1.23.0/or-tools_universal_macOS-14.4.1_go_v9.10.4129.tar.gz
to your /usr/local/lib folder to run this project
` go run cmd/main.go`

# Test build for amd64 linux systems
`podman buildx build --platform linux/amd64 -t myapp:latest --load .`
I was unable to build a stack binary
