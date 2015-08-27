package main

import (
  "io"
  "log"
  "net"
  "crypto/rand"
  "crypto/tls"
)

func echo(con net.Conn) {
    _, err := io.Copy(con, con)
    if err != nil {
        log.Print(err)
    }
    err = con.Close()
    if err != nil {
        log.Print(err)
    }
}

func main() {
  log.SetFlags(log.Lshortfile)

  cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
  if err != nil {
    log.Fatal(err)
  }

  config := tls.Config{Certificates: []tls.Certificate{cert} }
  config.Rand = rand.Reader

  ln, err := tls.Listen("tcp", ":10101", &config)
  if err != nil {
    log.Fatal(err)
  }
  for {
    con, err := ln.Accept()
      if err != nil {
        log.Fatal(err)
      }
    go echo(con)
  }
  err = ln.Close()
  if err != nil {
    log.Fatal(err)
  }
}
