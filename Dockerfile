# Base image setup with required tools and libraries
FROM quay.io/pypa/manylinux2014_x86_64:latest AS env

# Install necessary development tools and dependencies
RUN yum -y update \
    && yum -y groupinstall 'Development Tools' \
    && yum -y install wget curl \
    pcre2-devel openssl \
    which redhat-lsb-core \
    pkgconfig autoconf libtool zlib-devel glibc-static glibc-devel \
    && yum clean all \
    && rm -rf /var/cache/yum

ENTRYPOINT ["/usr/bin/bash", "-c"]
CMD ["/usr/bin/bash"]

# Install Go 1.23.0
RUN wget -q --no-check-certificate "https://go.dev/dl/go1.23.0.linux-amd64.tar.gz" \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz \
    && rm go1.23.0.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
RUN GOBIN=/usr/local/go/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33
RUN go version

# OR-TOOLS Installation
FROM env AS devel
WORKDIR /root/or-lib
RUN wget -q https://github.com/AirspaceTechnologies/or-tools/releases/download/v9.10-go1.23.0/or-tools_x86_64_CentOS-7.9.2009_go_v9.10.4129.tar.gz \
    && tar -xf or-tools_x86_64_CentOS-7.9.2009_go_v9.10.4129.tar.gz --strip 1

FROM devel AS builder
WORKDIR /root
COPY . .

# Ensure library paths are available for both build and runtime
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-L/root/or-lib"
ENV LD_LIBRARY_PATH=/root/or-lib:$LD_LIBRARY_PATH
RUN go build -a -ldflags '-linkmode external' -o test_air ./cmd

# Final image to run the application
FROM quay.io/pypa/manylinux2014_x86_64:latest AS deploy
COPY --from=builder /root/test_air ./
ENV LD_LIBRARY_PATH=/root/or-lib:$LD_LIBRARY_PATH
COPY --from=builder /root/or-lib /root/or-lib 
RUN ls 
RUN /test_air
