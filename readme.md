# expect implementation in go

- this implementation should mimic a simple flow for expect extension to automate interacting with cli programs.
- it cover the implementation for (spawn/expect/send). it should be compilable with any expect script uses only those methods.
- more about [expect](https://en.wikipedia.org/wiki/Expect)
- the main idea in this implementation is utilizing linux [pipes](https://man7.org/linux/man-pages/man2/pipe.2.html)

## Usage

- a simple bash script mimics a process to interact with.

  ```bash
  #!/bin/bash

  echo username
  read username
  echo password
  read password

  echo $username:$password > cred.txt
  ```

- the expect script that automate interacting with the process.

  ```exp
  #!/usr/local/bin/myexpect

  spawn ./ask
  expect "username"
  send "omar1"
  expect "password"
  send "pass1"
  expect eof
  ```

- run the app with the expect script

  ```bash
  go run main.go answer.exp
  ```

## Implementation

> TODO: explain
