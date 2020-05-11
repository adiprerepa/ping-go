# golang-ping
This a ping program written using golang (1.13). The project uses no external libraries besides the icmp
golang library, which is: `golang.org/x/net`.


This library is necessary for making raw ICMP calls.
## Project Structure

- The two sections are `pkg/agent` and `main.go`. `pkg/agent` holds driver and utility functions necessary for 
making the echo requests. It also holds unit tests for all the code. `main.go` holds the 
actual implementation of the interface to `pkg/agent`.
- The code also holds unit tests (45) to cover a vast majority of the functions found
in `pkg/agent`.

## Running/Execution
In order to run the actual program, you can either execute the pre-built binary in `/bin`,
or build the program with:
```
make build
```

You Can run the program with:
```shell script
sudo ./bin/ping
```
Then, a list of flags and usage will be given to you. Note: The program needs to be run in `sudo`
mode because golang needs to send ICMP packets.

## Techincal Concepts/Design

- The Program uses goroutines, which are basically threads, to receive packets and listen for 
Keyboard interrupts (ctrl+c).
- `ping_agent.go` actually holds the logic of starting/terminating goroutines, sending/receiving
ICMP packets. It also holds the data/pinger structs and status/statistics callbacks.
- `options.go` holds the implementation of parsing command line arguments, and putting 
up safeguards to keep corrupted/invalid data from entering the program.
- `util.go` holds utility functions that would not be in place otherwise.
- We send the time of sending and a tracker in every ICMP packet to track the packets. Time
to live is a custom option that is by default 255.
- Ipv6 Support is included.
