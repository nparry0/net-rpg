#!/bin/bash

openssl req -out ca.pem -new -x509 
openssl genrsa -out server.key 1024 
openssl req -key server.key -new -out server.req 
echo 00 > file.srl
openssl x509 -req -in server.req -CA ca.pem -CAkey privkey.pem -CAserial file.srl -out server.pem 
