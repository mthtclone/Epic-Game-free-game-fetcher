## Free Epic Game Fetcher

Free Epic Game Fetcher is a lightweight tool that checks the Epic Games Store for currently free games and notifies you automatically.

The program comes in two flavor: CLI & Systray, which you download from [release](https://github.com/mthtclone/Epic-Game-free-game-fetcher/releases).

For CLI, you may use it in extension with programmatic client. It supports JSON, HTML, and normal text for logging and automation. 

Flags:

- `--output=<file path>`  
  Write output to a file instead of stdout.

- `--append`  
  Append output to the specified file instead of overwriting it.
(Use this flag if you do not want to log mulitple files)

- `--format=<text|json|html>`  
  Explicitly specify the output format.  
  If omitted, the format is inferred from the output file extension.

Clone the repo, and run the following line to compile the program. 
```
go build -ldflags "-H windowsgui" -o freeEpicWatcher.go main.go output.go
```
and run the program: 
```
./freeEpicWatcher.go
