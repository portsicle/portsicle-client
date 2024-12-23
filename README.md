## Portscile Client

- Portsicle is a free and open-source Ngrok alternative to expose local servers online.

- Portsicle client allows you to use the <a href="github.com/portsicle/portsicle-server">Portsicle Server</a> via CLI.

## Installation guide

- Install the provided binary from latest release.

- Give executeable permission to the binary `chmod +x ./portsicle-client`.

- Use the CLI to run the client:

```
./portsicle http -p 3000
```

> note that `3000` is the port of your local server, which you want to expose on public network.

## CLI Usage

```
./portsicle --help

Usage:
  portsicle [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  http        Expose local http port

Flags:
  -h, --help     help for portsicle
  -t, --toggle   Help message for toggle

Use "portsicle [command] --help" for more information about a command.
```

```
./portsicle http --help
Expose local http port

Usage:
  portsicle http [flags]

Flags:
  -h, --help          help for http
  -p, --port string   Port on which your local server is listening. (default "8888")
```
