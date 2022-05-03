# pom

*pom* is a command line [pomodoro](https://en.wikipedia.org/wiki/Pomodoro_Technique) timer with vim keybindings, scriptable hooks and json logging.

![image](https://user-images.githubusercontent.com/173962/166444425-c2a81732-41f7-40c2-bf11-f631f38948b4.png)

<!-- vim-markdown-toc GFM -->

* [installation](#installation)
* [usage](#usage)
* [configuration](#configuration)
  * [hook profiles](#hook-profiles)
  * [show status in tmux](#show-status-in-tmux)

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
  * `Esc` clear a chord
  * `hjklgG` move around
  * `cc` *continue* with next break/pomodoro
  * `q` quit the program
* *pom* logs all completed pomodoros to: `~/.pom/log.json`

```json
[
  {
    "CurrentTask": "research",
    "PomDuration": 60000000000,
    "StartTime": "2022-05-03T13:07:40.091129279+02:00",
    "State": "work_done",
    "StopTime": "2022-05-03T13:08:40.285970461+02:00"
  },
...
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
* when *pom* exits this file is emptied

You can `cat` this into your tmux statusline:

**.tmux.conf**
```sh
set -g status-right "[#(cat ~/.pom/tmux)]"
```
