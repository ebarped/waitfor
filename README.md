# waitfor
Wait until a system:port is up while you mind your own business... 

## Prerequisites
https://github.com/ebitengine/oto#prerequisite

```bash
sudo apt install libasound2-dev
```

## Use Case
A lot of times you need to wait for a host machine to be up...
e.g: Create an instance in the cloud and wait until the ssh port is up so you can connect.
Instead of trying several times to connect, or use **netcat** to warn you, you can use **waitfor**!

**waitfor** will try to connect to the `host:port`, with a default timeout of **10m** (configurable). Wheter it success or fails, **waitfor** will send you a nice OS notification with the result, and a wake-up sound.

## Usage
```bash
waitfor google.com:80
```

Specifying timeout value (valid time units are "s", "m", "h")
```bash
waitfor -timeout 1m 8.8.8.8:53 
```

## Crosscompile for windows/darwin
We use github.com/crazy-max/xgo to crosscompile with CGO and additional compilers.
```
xgo '-targets=windows/*,darwin/*' .
```

## TODOS
- use multiple regexps (1 for scheme://host:port, another for host:port, another for ip:port)
- test in windows & mac
