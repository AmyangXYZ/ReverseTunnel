# Reverse Tunnel

A simple realization of SSH reverse tunnel.

Just establish a tunnel between master and slave, and forward all requests and response to it. Like a reverse shell that only does network requests.

## Usage

The commands below equal to `ssh -R 0.0.0.0:10002:127.0.0.1:22 root@192.168.3.79`

### Master

compile.

`$ go build -o master master.go tunnel.go`

tunnel port 10001, public port 10002.

`$ ./master -t 0.0.0.0:10001 -l 0.0.0.0:10002`

### Slave

compile.

`$ go build -o slave slave.go tunnel.go`

tunnel addr 192.168.3.79:10001, connect to localhost:22

`$ ./slave -t 192.168.3.79:10001 -c localhost:22`