# GompherMP

An implementation of parallelism in the Go programming language using OpenMP-like clauses.

## About This Project
This repository contains the ongoing development and documentation for GompherMP. The project focuses on adapting a subset of OpenMP directives to Go. The goal is to allow for explicit and structured concurrent execution using Go's native goroutines and synchronization primitives.

## Repository Structure
Currently, the repository is organized to house the formal documentation:
* `/docs/specs`: Contains the technical specification of the GompherMP directives, formal syntax, and expected runtime behaviors.
* `/docs/thesis`: Contains the drafts and chapters for the final thesis document.

## Setup for Developers

1. **Prerequisites**: Install Go 1.22.2.
2. **Environment**: We use a `Makefile` to simplify tasks.
3. **First time setup**:
    ```bash
    make deps
    ```
4. **To Build**:
    ```bash
    make build
    ```
    This creates a ./gompher executable.
5. **To Test:**:
    ```bash
    make test
    ```
## Authors
* Jorge David Alejandro Contreras
* Patricia Natividad Cántaro Márquez
