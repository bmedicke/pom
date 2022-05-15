# pom

*pom* is a command line [pomodoro](https://en.wikipedia.org/wiki/Pomodoro_Technique)
timer with vim keybindings, scriptable hooks, a web API, tmux support and json logging.

![image](https://user-images.githubusercontent.com/173962/166680550-d70ed16a-bc93-414e-bf04-ad42abcf9f96.png)

<!-- vim-markdown-toc GFM -->

* [installation](#installation)
* [usage](#usage)
* [flags](#flags)
* [configuration](#configuration)
  * [hooks](#hooks)
  * [hook profiles](#hook-profiles)
  * [show status in tmux](#show-status-in-tmux)
  * [JSON logging](#json-logging)
  * [Web API](#web-api)

<!-- vim-markdown-toc -->

## installation

```sh
go install github.com/bmedicke/pom@latest
```

## usage

```sh
pom
```

* **keyboard shortcuts** are loosely based on Vim
  * `hjklgG` move around
  * `a`/`A`/`Enter` append to cell
  * `cc` change cell
  * `dd`/`dc` delete cell content
  * `Esc` clear a key chord
  * `;` next break/pomodoro
  * `q` quit the program
  * `Q` quit the program (don't save incomplete pomodoros)
* while editing a cell [readline keybindings](https://en.wikipedia.org/wiki/GNU_Readline) are available
  * `ctrl-a` jump to beginning
  * `ctrl-e` jump to end
  * `ctrl-k` delete to the right of cursor
  * etc.

## flags

* `-h` show the help
* `--profile <subdir>` select non-default [hook profile](#hook-profiles)
* `--create-config` create config files, see [next section](#configuration)
* `--longbreak-in <uint>` overwrite number of pomodoros required for the first long break


## configuration

Create the config folder, config file and default hooks at `~/.config/pom/`:

```sh
pom --create-config
```

### hooks

* edit the scripts in `~/.config/pom/hooks/default`
* the scripts are named after when they are called:
  * `work_start`
  * `work_done`
  * `break_start`
  * `break_done`
  * `longbreak_start`
  * `longbreak_done`
  * `pomodoro_cancelled`
* the interpreter (`sh`, `zsh`, `python3`, etc.) of the script is set via the [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix))

Here are a couple of **usage scenarios** for hooks:

  * start/stop your music (e.g. via `ncmpcpp`/`mpc`)
  * toggle a light (e.g. via Home Assistant's `hass`)
  * send a notification (via `prowl`'s REST API, etc.)
  * set a Home Assistant `input_boolean` for further scripting

### hook profiles

* the `default` hooks profile is used when no other is specified
* you can create custom hook profiles
* e.g. a profile that toggles a light when pomodoros/breaks end:
  * create a folder `~/.config/pom/hooks/light`
  * add a `work_done` and a `break_done` script to it
  * add your shell commands to toggle the light
  * start *pom* with your profile:

```sh
pom --profile light
```

### show status in tmux

* if `writeTmuxFile` in `~/.config/pom/config.json` is set to `true`:
  * *pom* keeps a file at `~/.config/pom/tmux` that always shows the current status
  * when *pom* exits this file is emptied

You can `cat` this file into your tmux statusline:

**.tmux.conf**
```sh
set -g status-right "[#(cat ~/.config/pom/tmux)]"
```

### JSON logging

* if `logJSON` in `~/.config/pom/config.json` is set to `true`:
  * *pom* logs all complete and incomplete pomodoros to: `~/.config/pom/log.json`

**~/.config/pom/log.json**

```json
[
  {
    "Project": "master thesis",
    "Task": "research",
    "Note": "mode locking",
    "State": "work_done",
    "Duration": 60000000000,
    "StartTime": "2022-05-03T13:07:40.091129279+02:00",
    "StopTime": "2022-05-03T13:08:40.285970461+02:00"
  },
...
```

### Web API

* disabled by default, edit `~/.config/pom/config.json`'s `enableAPI` to activate this feature
* the API is still quite rudimentary
* there are two endpoints that respond with JSON
  * GET `/continue`: starts the next state (same as pressing `;`)
  * GET `/ws`: upgrades connection to a websocket and streams infos about the current pomodoro twice a second
  * GET `/live`: open in browser to have a tab with current timestamp in the title

```sh
websocat ws://localhost:8421/ws | jq
```

```json
{
  "Project": "pom",
  "Task": "hide server from table if disabled",
  "Note": "go lang",
  "Duration": 1500000000000,
  "StartTime": "2022-05-07T14:23:36.755337148+02:00",
  "State": "work",
  "StopTime": "0001-01-01T00:00:00Z"
}
...
```

```sh
curl localhost:8421/continue # or call it from a bookmark, your phone, etc.
```
