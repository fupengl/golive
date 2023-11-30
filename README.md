# Go live

GoLive is a lightweight Go language hot reload tool designed to simplify the development process and improve development efficiency.    
By monitoring project files for changes, GoLive automatically detects and restarts your Go services.

# Installation

Make sure you have [Go](https://go.dev/) installed on your machine. Then, run the following command to install GoLive:

```shell
go install github.com/fupengl/golive@latest
```

# Usage

To use GoLive, replace the usual `go run` command with `golive`. For example:

```shell
golive main.go
```

This will start your Go application and automatically restart it whenever there are changes to the source code.

# How It Works

GoLive utilizes Go modules to discover and monitor changes in your project. It automatically finds the working directory and detects any replace directives declared in your go.mod file.

1. Working Directory: GoLive automatically identifies the working directory based on your project structure.

2. Go Modules: It leverages the information from your go.mod file, ensuring that replace directives and module dependencies are correctly handled.

3. Automatic Restart: Whenever there are changes detected in the source code, GoLive intelligently restarts your Go application without manual intervention.

# Features

- Automatic restart on file changes
- Simplifies the development process
- Easy to integrate into your existing projects

# License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.