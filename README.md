# Shared Library Project

> This shared library provides various utility functions and structures that can be utilized across different Go
> projects. The goal of this library is to promote code reuse and to offer a central repository of common code
> components.

<!-- TOC -->
* [Shared Library Project](#shared-library-project)
  * [Features](#features)
  * [Linting](#linting)
  * [Testing](#testing)
  * [Contributing](#contributing)
  * [License](#license)
<!-- TOC -->

## Features

This shared library includes a range of features designed to simplify development across various projects:

* **Logging**: A versatile logging package that supports different log levels and output formats.
* **Networking Utilities**: Helper functions and structures for common networking tasks, including HTTP and WebSocket
  communication.
* **Service Discovery**: Implementations for mDNS and SSDP service discovery protocols.
* **Command-Line Interface**: Utilities to assist in building CLI applications, including version and license commands.

Each feature is encapsulated in its own module within the library, providing a modular approach to using specific
functionalities.

## Linting

Ensure the quality of the repository met my standards:

```bash
task golangci
```

## Testing

To test the shared library, navigate to the root directory of the library and run:

```bash
task go:test
```

Ensure that all tests pass before integrating the library into your project.

## Contributing

Contributions are welcome! If you would like to contribute, please:

1. Fork the repository.
2. Create a new branch for your changes.
3. Make your changes and write any necessary tests.
4. Ensure that all tests pass.
5. Submit a pull request to the main repository.

## License

This shared library is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.
