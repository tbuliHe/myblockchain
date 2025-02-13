# A Simple Modular Blockchain in Go

## Overview

A lightweight blockchain implementation with basic features written in Go.


Based on [@anthdm's](https://github.com/anthdm) blockchain tutorial series.
Original repo: https://github.com/anthdm/projectx

## Features

- **Basic Blockchain Structure**
  - Block creation and validation
  - Transaction signing and verification
  - Transaction mem pool
  - Simple consensus mechanism

- **Network Layer**
  - Local P2P network simulation
  - Transaction broadcasting
  - Block propagation
  - TCP transport
  - JSON server

- **Virtual Machine**
  - Simple stack-based VM
  - Basic instruction set
  - State management

- **Cryptography**
  - ECDSA for digital signatures
  - SHA256 for hashing
  - Public/private key management

## Getting Started

### Prerequisites
- Go 1.20 or higher

### Installation
```bash
git clone <repository-url>
cd myblockchain
make build
```

### Running Tests
```bash
make test
```

### Running the Node
```bash
make run
```

