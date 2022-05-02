# pom

a command line pomodoro timer

## installation

```sh
go install github.com/bmedicke/bhdr@latest
```
## configuration

* create the config structure with the `--create-config` flag
* edit the shell scripts in `~/.pom/callbacks`
* the scripts are named after when they are called:
  * `work_start.sh`
  * `work_done.sh`
  * `break_start.sh`
  * `break_done.sh`
