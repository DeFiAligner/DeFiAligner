FROM ubuntu:22.04
#FROM ghcr.io/z3prover/z3:ubuntu-20.04-bare-z3-sha-e7c17e6


ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get -y --no-install-recommends install \
    cmake \
    make \
    clang \
    g++ \
    curl \
    default-jdk \
    python3 \
    python3-setuptools \
    python3-pip \
    python-is-python3 \
    sudo \
    golang-go \
    git && \
    apt-get clean && rm -rf /var/lib/apt/lists/*


# Install Z3
RUN git clone https://github.com/Z3Prover/z3.git /z3 && ls /z3
# Run mk_make.py to prepare the build directory
WORKDIR /z3
RUN python scripts/mk_make.py --python
# Build and install Z3
WORKDIR /z3/build
RUN make
RUN sudo make install


COPY . /DeFiAligner-v1.0
WORKDIR /DeFiAligner-v1.0

# # Install go-z3
# WORKDIR /DeFiAligner-v1.0/go-z3
# RUN go mod vendor
# RUN make



# Reset the work directory if necessary
WORKDIR /

ENTRYPOINT ["/bin/bash"]
