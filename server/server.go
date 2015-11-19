package main

import (
  "log"
  "github.com/nparry0/network"
)

var gWorld *World
var gClients []*ClientConn

func main() {
  var err error
  log.SetFlags(log.Lshortfile)

  log.Printf("Initializing world\n"); 
  gWorld, err = NewWorld()
  if err != nil {
    log.Printf("FAILED\n"); 
    log.Panic(err)
  }
  log.Printf("Done %v\n", gWorld.Regions["Test Chambers"]); 

  ln, err := network.Listen("")
  if err != nil {
    log.Fatal(err)
  }

  for {
    conn, err := network.Accept(ln)
    if err != nil {
      log.Print(err)
    } else {
      // TODO: Clean up when the client disconnects.
      // TODO: Better synchronization on gClients.
      gClients = append(gClients, NewClientConn(conn)) 
    }
  }

  err = ln.Close()
  if err != nil {
    log.Fatal(err)
  }
}
