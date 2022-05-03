# pom

a command line pomodoro timer


<!-- vim-markdown-toc GFM -->

* [installation](#installation)
* [configuration](#configuration)
  * [hook profiles](#hook-profiles)

<!-- vim-markdown-toc -->

## installation

```sh
go install github.com/bmedicke/bhdr@latest
```
## configuration

* create the config folders and default hooks with the `--create-config` flag
* edit the shell scripts in `~/.pom/hooks/default`
* the scripts are named after when they are called:
  * `work_start.sh`
  * `work_done.sh`
  * `break_start.sh`
  * `break_done.sh`

### hook profiles

* the default hooks profile is used when no other is specified
* you can create custom profiles
* e.g. one that toggles a light when pomodoros/breaks end:
  * create a folder `~/.pom/hooks/light`
  * add a `work_done.sh` and a `break_done.sh` script to it
  * add your shell commands to toggle the light
  * start *pom* with your profile: `pom --profile light`
