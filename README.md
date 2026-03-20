# ssh-connect

`ssh-connect` is a simple CLI tool to manage multiple SSH accounts and named servers from one JSON config file.

## Features

- `ssh-connect to <servername>` connect using server + account mapping
- `ssh-connect list` show all accounts and servers
- `ssh-connect list account` show account list
- `ssh-connect add account` add account (interactive or via flags)
- `ssh-connect delete account` delete account (interactive or via arg/flag)
- `ssh-connect list server` show server list
- `ssh-connect add server` add server (interactive or via flags)
- `ssh-connect delete server` delete server (interactive or via arg/flag)
- `ssh-connect passphrase account` set passphrase for SSH account
- `ssh-connect setup` setup permission and auto-register accounts from `.ssh`

## Config

Default config path is `~/.ssh-connect/ssh-connect-config.json`.

You can override it with:

```sh
SSH_CONNECT_CONFIG=/path/to/ssh-connect-config.json ssh-connect list
# or
ssh-connect --config /path/to/ssh-connect-config.json list
```

Example config format:

```json
{
    "accounts": [
        {
            "name": "team-dev",
            "email": "dev@example.com",
            "path": "~/.ssh/team-dev/id_ed25519",
            "passphrase": ""
        }
    ],
    "servers": [
        {
            "name": "prod-1",
            "username": "root",
            "host": "10.10.10.10",
            "port": 22,
            "default_account": "team-dev"
        }
    ]
}
```

## Usage

```sh
# list
ssh-connect list
ssh-connect list account
ssh-connect list server

# add using interactive prompt
ssh-connect add account -i
ssh-connect add server -i

# add server: default account input now suggests existing accounts
# and will keep asking until valid (no immediate exit on wrong value)

# add using parameters
ssh-connect add account --name team-dev --email dev@example.com --path ~/.ssh/team-dev/id_ed25519
ssh-connect add server --name prod-1 --username root --host 10.10.10.10 --port 22 --account team-dev

# set or clear account passphrase
ssh-connect passphrase account team-dev --value "your-passphrase"
ssh-connect passphrase account team-dev --clear

# delete interactive (if no name is provided)
ssh-connect delete account
ssh-connect delete server

# delete by name
ssh-connect delete account team-dev
ssh-connect delete server prod-1

# connect
ssh-connect to prod-1

# fix key permissions (Linux/macOS)
ssh-connect setup
```

## Build

```sh
make test
make windows
make macos
make deb
make all
```

When installed from the Debian package, shell completion files are generated automatically during install for Bash, Zsh, and Fish.
