FROM ubuntu:24.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install essential packages
RUN apt-get update
RUN apt-get install -y \
    build-essential \
    curl \
    git \
    g++ \
    gcc \
    make \
    libxml2-dev \
    libxml2-utils \
    pkg-config \
    software-properties-common \
    sudo \
    unzip \
    vim \
    wget \
    zip \
    && curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget https://golang.org/dl/go1.25.6.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.25.6.linux-amd64.tar.gz \
    && rm go1.25.6.linux-amd64.tar.gz

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/home/developer/go"
ENV GOROOT="/usr/local/go"

# Create developer user and set up workspace
RUN useradd -m -s /bin/bash developer \
    && mkdir -p /workspace \
    && chown -R developer:developer /workspace \
    && chown -R developer:developer /home/developer

# Create GOPATH directories before switching user
RUN mkdir -p /home/developer/go/{bin,src,pkg,pkg/mod} \
    && chown -R developer:developer /home/developer/go

# Install additional development tools
USER developer
WORKDIR /workspace

# Create GOPATH directories
RUN mkdir -p $GOPATH/{bin,src,pkg,pkg/mod}

# Add Go bin to PATH for the developer user
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

# Expose ports commonly used by the application
EXPOSE 3000 8080 9090

CMD ["/bin/bash"]