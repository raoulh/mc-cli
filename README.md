Moolticute CLI tool
====================

[![License](https://img.shields.io/badge/license-GPLv3%2B-blue.svg)](http://www.gnu.org/licenses/gpl.html)

This tool is a command line tool to interact with a running moolticute daemon. It allows to retrieve credentials from a CLI.

### Installation ###

Install go if needed. Add `$GOPATH/bin` to your path.

```
go get github.com/raoulh/mc-cli
```

### Usage ###

```
mc-cli --help

Usage: mc-cli COMMAND [arg...]

Command line tool to interact with a mooltipass device through a moolticute daemon

Commands:
  login        Manage credentials stored in the device
  data         Import & export small files stored in the device
  parameters   Get/Set device parameters

Run 'mc-cli COMMAND --help' for more information on a command.
```

```
mc-cli login --help

Usage: mc-cli login COMMAND [arg...]

Manage credentials stored in the device

Commands:
  get          Get a password for given context
  set          Add or update a context

Run 'mc-cli login COMMAND --help' for more information on a command. A typical example is

  /mc-cli login get github.com ""

or to fetch the password for your github account. Add the flag '-l' to also extract the username:

  /mc-cli login get github.com "" -l

Alternatively use:

  /mc-cli login get github.com myusername

if you want to be specific about a username.
```

```
mc-cli data --help                                                                                                               16:45:08  ✘ 2

Usage: mc-cli data COMMAND [arg...]

Import & export small files stored in the device

Commands:
  get          Retrieve data for given context
  set          Add or update data for given context

Run 'mc-cli data COMMAND --help' for more information on a command.
```
> Warning! This project is a work in progress!
