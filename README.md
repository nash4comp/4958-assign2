## COMP4958 Assignment 2
### Nash Baek, A01243888

### How to run
#### Server
within server folder
> go run server.go

#### Client
> telnet localhost 6666

within client folder
> java Main.java

### How to Terminate client
for telnet connection
> ctrl + ], quit

other connection
> ctrl + c

### Commands List
- /LIST
  - This command lists only nickname of the clients who are registered nickname before calling this command.
- /NICK <nickname>
- /BC <message>
  - This command broadcasts to only nickname of the clients who are registered nickname before calling this command.
- /MSG <nickname> <message>