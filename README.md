Moolticute CLI tool
====================

[![License](https://img.shields.io/badge/license-GPLv3%2B-blue.svg)](http://www.gnu.org/licenses/gpl.html)

This tool is a command line tool to interact with a running moolticute daemon. It allows to retrieve credentials from a CLI.

### Installation ###

Install go if needed. Add `$GOPATH/bin` to your path.

```
go get github.com/raoulh/moolticute-cli
```

### Usage ###

```
moolticute-cli --help

Usage: moolticute-cli COMMAND [arg...]

Command line tool to interact with a mooltipass device through a moolticute daemon

Commands:
  login        Manage credentials stored in the device
  data         Import & export small files stored in the device
  parameters   Get/Set device parameters

Run 'moolticute-cli COMMAND --help' for more information on a command.
```

```
moolticute-cli login --help

Usage: moolticute-cli login COMMAND [arg...]

Manage credentials stored in the device

Commands:
  get          Get a password for given context
  set          Add or update a context

Run 'moolticute-cli login COMMAND --help' for more information on a command.
```

```
moolticute-cli data --help                                                                                                               16:45:08  ✘ 2

Usage: moolticute-cli data COMMAND [arg...]

Import & export small files stored in the device

Commands:
  get          Retrieve data for given context
  set          Add or update data for given context

Run 'moolticute-cli data COMMAND --help' for more information on a command.
```
> Warning! This project is a work in progress!
