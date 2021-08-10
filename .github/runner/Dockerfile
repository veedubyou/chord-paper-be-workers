FROM tcardonne/github-runner:latest

# install lib-ssl to fix problem with missing OpenSSL
RUN apt-get update && \
    apt-get install -y \
        libssl-dev \
        pkg-config
