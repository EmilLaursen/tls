# How to install

The binary can be build with
```
make tls-linux
make tls-osx
make tls-windows
``` 
and then move the binary on to your PATH. It does not require golang to be installed, but it does require docker.
I have no idea whether the osx and windows builds work.

# TODO

- [x] verbose flag showing progress bar, and speed in MiB/sec
- [ ] bounded concurrency
- [ ] web api