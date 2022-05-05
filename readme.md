# pom

*pom* is a command line [pomodoro](https://en.wikipedia.org/wiki/Pomodoro_Technique) timer with vim keybindings, scriptable hooks, tmux support and json logging.

![image](https://user-images.githubusercontent.com/173962/166680550-d70ed16a-bc93-414e-bf04-ad42abcf9f96.png)

<!-- vim-markdown-toc GFM -->

* [installation](#installation)
* [usage](#usage)
* [configuration](#configuration)
  * [hooks](#hooks)
  * [hook profiles](#hook-profiles)
  * [show status in tmux](#show-status-in-tmux)
  * [JSON logging](#json-logging)

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
  * `q`/`Q` quit the program

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
