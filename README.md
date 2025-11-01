# gossher

Infrastructure management tool with GUI, powered by GO.

## Overview

**gossher** is a modern infrastructure management tool. Manage servers, SSH credentials, and host groups efficiently through an intuitive GUI interface.

## Features

- 🖥️ **Cross-platform GUI** - Built with [Gio UI](https://gioui.org)
- 📝 **YAML-based Configuration** - Easy to read and edit
- 🔐 **SSH Credential Management** - Secure storage for SSH keys and passwords
- 🏗️ **Host Organization** - Group and manage hosts collectively
- 📊 **Flexible YAML Structure** - Define multiple entities in single files

## Tech Stack

| Component | Technology | License          |
|-----------|-----------|------------------|
| **GUI Framework** | [Gio UI](https://gioui.org) | MIT / UNLICENSE  |
| **Language** | Go 1.25+ | -                |
| **Configuration** | [YAML v3](https://github.com/go-yaml/yaml) | Apache 2.0 / MIT |
| **SSH Operations** | [Go Crypto](https://golang.org/x/crypto) | BSD 3-Clause     |

## Installation

### Prerequisites

- Go 1.25 or later
- CGO enabled (required for Gio UI)

### From Source

