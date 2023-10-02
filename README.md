# UpdateHostsFile

```
 _   _           _       _       _   _           _       _____ _ _
| | | |_ __   __| | __ _| |_ ___| | | | ___  ___| |_ ___|  ___(_) | ___
| | | | '_ \ / _  |/ _  | __/ _ \ |_| |/ _ \/ __| __/ __| |_  | | |/ _ \
| |_| | |_) | (_| | (_| | ||  __/  _  | (_) \__ \ |_\__ \  _| | | |  __/
 \___/| .__/ \__,_|\__,_|\__\___|_| |_|\___/|___/\__|___/_|   |_|_|\___|
      |_|
```

`UpdateHostsFile` (or `update-hosts-file`) is a program written in Go that automates the task of updating the /etc/hosts file on Unix-based systems. It can be configured to pull host information from various sources such as web-based and local blocklists files. It also automatically adds the hostname of the machine to make sure any changes to the hostname will be reflected in the /etc/hosts file.

# Building the Program

> I made available in this repo a PKGBUILD to make it possible to build and install easily on Arch Linux

To build the program, make sure that Go is installed on your system. Clone the repository or download an archive for a specific version and run the following command in the terminal:

```bash
make build
```

This will create a binary file called `update-hosts-file`, autocompletion files for `bash`, `zsh`, and `fish` in the directory `./autocompletions`, and markdown documentation files in the directory `./out`. To install the binary and program files (including the default modules, default `preferences` file, autocompletion files, and the documentation):

```bash
sudo make install
```

And to uninstall:

```bash
sudo make uninstall
```

# Post-installation

After installing, run:

```bash
sudo update-hosts-file modules enable --local --module default
```

To enable the `default` local module, which contains the default hosts for Unix-based systems.

# Documentation

## Modules

The `update-hosts-file` program allows the user to update their /etc/hosts file using local and web modules. In the case of local modules, the file is parsed directly and appended to a temporary hosts file. On the other hand, in the case of web modules, the hosts file is downloaded, parsed and appended to the temporary file. Finally, this temporary hosts file is moved to /etc/hosts.

### Local

Local modules use the same syntax as the /etc/hosts file and are located at `/usr/share/update-hosts-file/modules/local`.

### Web

Web modules files only contain the URL for the source of the hosts file and are located at `/usr/share/update-hosts-file/modules/web`.

### Enabling/Disabling

To enable a module, the program links it from the available directory to the enabled directory.

For example, to enable the `default` local module, the program will create a symbolic link from `/usr/share/update-hosts-file/modules/local/available/default` to `/usr/share/update-hosts-file/modules/local/enabled/default`. This method is similar to the one used by Apache web server.

And to disable it, the program removes this symbolic link.

## Preferences

The program includes a configuration file that allows you to customize its behavior. The file is located at `/usr/share/update-hosts-file/config/preferences` and it follows this format:

```bash
VAR1=VALUE
VAR2=VALUE
```

Here's a description of each configuration variable:

- `DEFAULT_EDITOR`: This variable sets the default editor to be used by the program when editing files. The value should be the full path to the desired editor executable, such as /usr/bin/nano, /usr/bin/vim, or /usr/bin/emacs. The default value is `/usr/bin/nano`.
- `DEFAULT_VIEWER`: This variable sets the default viewer to be used by the program when displaying files. The value should be the full path to the desired viewer executable, such as /usr/bin/less, /usr/bin/cat, or /usr/bin/batcat. The default value is `/usr/bin/cat`.
- `MAX_BACKUP_FILES`: This variable sets the maximum number of backup files that the program will keep. Before overwriting the /etc/hosts file, a backup is created in the backup directory. If the number of backup files in the directory exceeds the value of this variable, the oldest backup files will be deleted. The default value is `10`.
- `KEEP_ON_HOST_UNREACHABLE`: This variable determines whether the program should skip a module and not restore its backup if the source of a web module cannot be reached. If the value is set to true, the program will finish with an error and the backup will be restored. If the value is set to false, the program will skip the module and keep loading other modules, if any. The default value is `false`.
- `IP_TEST`: This variable sets the IP address or hostname of a remote server that the program will use to test the internet connection. The program will attempt to ping the server and check for a response. If no response is received, the program will assume that the internet connection is down and will not attempt to download any updates. The default value is `8.8.8.8`, which is a public DNS server operated by Google.

## Available Subcommands

`update-hosts-file enable`

This subcommand enables the systemd service on boot

`update-hosts-file disable`

This subcommand disables the systemd service on boot

`update-hosts-file version`

This subcommand displays the version of the program.

`update-hosts-file help`

This subcommand displays help information about the program.

`update-hosts-file modules`

This subcommand allows managing the modules used to update the /etc/hosts file. The following subcommands are available:

- `enable`: enables a module
- `disable`: disables a module
- `add`: adds a new module
- `rm`: removes an existing module
- `edit`: edits an existing module
- `view`: views the content of an existing module
- `list`: list existing modules and show if they are enabled or disabled

### Examples

Enabling a web module

```bash
sudo update-hosts-file modules enable --web --module example_module
```

Disabling a local module

```bash
sudo update-hosts-file modules disable --local --module example_module
```

Adding a module

```bash
sudo update-hosts-file modules add --local --module example_module
```

Editing a module

```bash
sudo update-hosts-file modules edit --local --module example_module
```

Viewing a module

```bash
sudo update-hosts-file modules view --local --module example_module
```

Listing local modules

```bash
sudo update-hosts-file modules list --local
```

Listing web modules

```bash
sudo update-hosts-file modules list --web
```

Listing all modules

```bash
sudo update-hosts-file modules list --all
```

## License

UpdateHostsFile is licensed under the GPL-3.0 license.
