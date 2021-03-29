# Better-Bjobs-Go

A Go rewrite of my original python curses-based wrapper script for LSF _bjobs_ command.

Standard _bjobs_ output is not interactive, is not color highlighted, and is not
information dense, and I was frustrated with there being no default 'dashboard'
option for bjobs, and instead having to write elaborate shell pipes to monitor
the jobs I was interested in.

## Features

- Color highlight jobs based on their status (`RUN`/`DONE`/`EXIT`)
- Interactive interface so no need to rerun bjobs or use `watch`
- Show each job's maximum RAM usage compared to how much it was allocated
- Show how close each job is to its time-limit
- Show only a subset of the total jobs that match a project name
- Display in red and move to top of screen jobs that are approaching their
time or RAM limit
- Display count of pending jobs but don't list each of them
- Receive email notification when jobs have finished with information on how many
succeeded and how many exited
- Option to kill all jobs at once
- See jobs that are no longer visible with `bjobs -a` thanks to job caching on
exiting interface

## Screenshots

View of Better-Bjobs interface with buttons at bottom, and color-highlighted example jobs

![View of Better-Bjobs interface with example jobs](img/sch-normal.svg)


When a job approches its memory or time limit it is highlighted in red and brought to the top for attention

![View of better-bjobs interface with alert indicating a job approaching time limit](img/sch-normal-alert.svg)


## Usage

If `bj` is in the `PATH`, then it can be run with the `bj` command
in a spare terminal window or tmux pane.
This will show all non-pending jobs for the user.

To show only jobs from a certain project (the most useful use-case) add the
name of the project as an argument to the `bj` command.

for example to show all bjobs related to the _fq compression_ project:

```{bash}
bj "fq compression"
```

To set a project name when launching the jobs specify the project name as the
`-Jd` argument for the `bsub` relative to that project.

## Installation

### Binary Release

A compiled Ubuntu x86 binary is avaliable as a release, and this is the easiest
way to start using it straight away as the binary is self-contained.

To quickly install :

```{bash}
mkdir -p $HOME/bin/ # make a bin folder in home dir if one doesn't exist

# download the binary for linux
curl -Ls https://github.com/seanlaidlaw/Better-Bjobs-Go/releases/download/0.8/bj --output $HOME/bin/bj
chmod +x $HOME/bin/bj # allow executable

# add to path so we can run just `bj` from anywhere if bin not already
# this should be copied to .bashrc or .zshrc for persistance
export PATH=$HOME/bin:$PATH
```

### Compilation from source

Better-Bjobs can be compiled from source with the usual `go build` but
requires the [termui](https://github.com/gizak/termui/) go library as a dependency.

## Dependencies

As a wrapper script, the following command line tools are requried to be found
in the PATH
(although this is likely already the case if you're configured for using LSF):

- bjobs
- bkill
- mail
