package main

import (
    "io"
    "log"
    "net"
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
    ln, err := net.Listen("tcp", ":10101")
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
