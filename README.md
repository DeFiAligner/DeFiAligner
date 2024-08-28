# DeFiAligner


##  Introduction
DeFiAligner focuses on detecting inconsistencies in balance change symbolic formulas triggered by DeFi functions with identical ABIs. 
*Please note that this project only showcases the core code for path generation and symbolic extraction. For utilizing  Large Language Models in detection, refer to available online resources.*


## Overview

```
.
├── abiparser  # Generates symbolic inputs based on ABIs
│   ├── abi_parser.go
│   └── ABIs
│       └── erc20.abi.json
├── case_test.go  # Usage examples
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── pathfeat   # Symbolic feature extraction
│   ├── defi_features_extractor.go
│   ├── symbol_simplifier.go
│   └── token_balance_symbol.go
├── README.md
├── sevm       # Ethereum Virtual Machine supporting Z3 symbols
│   ├── common.go
│   ├── config.json
│   ├── contract.go
│   ├── execution_path.go
│   ├── instructions.go
│   ├── interpreter.go
│   ├── jump_table.go
│   ├── memory.go
│   ├── sevm.go
│   └── stack.go
└── tracer     # Used for debugging, compares generated symbolic paths with actual paths
    └── tracer.go

```

## Installation

1. Download the project code.
2. Build and run using Docker:

```
sudo docker build -t DeFiAligner-v1.0
sudo docker run -it --name test-container DeFiAligner-v1.0
```

## Usage

For usage examples, see  `case_test.go`.
