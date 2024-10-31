#!/bin/bash

echo username
read username
echo password
read password

echo $username:$password > cred.txt
