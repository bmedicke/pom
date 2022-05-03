# pom

a command line pomodoro timer


<!-- vim-markdown-toc GFM -->

* [installation](#installation)
* [configuration](#configuration)
  * [hook profiles](#hook-profiles)
  * [show status in tmux](#show-status-in-tmux)

<!-- vim-markdown-toc -->

## installation

```sh
go install github.com/bmedicke/pom@latest
```
## configuration

* create the config folders and default hooks with the `--create-config` flag
* edit the scripts in `~/.pom/hooks/default`
* the scripts are named after when they are called:
  * `work_start`
  * `work_done`
  * `break_start`
  * `break_done`
* the interpreter (`sh`, `zsh`, `python3`, etc.) of the script is set via the [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix))

Here are a couple of example **usage scenarios** for hooks:

  * start/stop your music (e.g. via `ncmpcpp`/`mpc`)
  * toggle a light (e.g. via Home Assistant's `hass`)
  * send a notification (via `prowl`'s REST API, etc.)
  * set a Home Assistant `input_boolean` for further scripting

### hook profiles

* the `default` hooks profile is used when no other is specified
* you can create custom profiles
* e.g. one that toggles a light when pomodoros/breaks end:
  * create a folder `~/.pom/hooks/light`
  * add a `work_done` and a `break_done` script to it
  * add your shell commands to toggle the light
  * start *pom* with your profile: `pom --profile light`

### show status in tmux

* *pom* keeps a file at `~/.pom/tmux` that always shows the current status

You can `cat` this into your tmux statusline:

**.tmux.conf**
```sh
set -g status-right "[#(cat ~/.pom/tmux)]"
```
