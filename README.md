# waitfor
Wait until a system:port is up while you mind your own business... 

## Use Case
A lot of times you need to wait for a host machine to be up...
e.g: Create an instance in the cloud and wait until the ssh port is up so you can connect.
Instead of trying several times to connect, or use **netcat** to warn you, you can use **waitfor**!

**waitfor** will try to connect to the `host:port`, with a default timeout of **10m** (configurable). Wheter it success or fails, **waitfor** will send you a nice OS notification with the result, and a wake-up sound.

## Usage 
```bash
waitfor google.com:80
```

```bash
waitfor 8.8.8.8:53 -timeout 1m
```

## Compatibility
As **waitfor** is written in go, is totally compatible with linux, mac & windows. Just download your OS/ARCH compatible binary from Releases

## TODOS
- flags
- progress bar
- notifications
- sound
- test in windows & mac
