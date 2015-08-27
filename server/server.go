package main

import (
  "log"
  "github.com/nparry0/network"
)

/* Each client gets one of these threads to listen to it,
   send data to it, and handle validation */
func clientConnThread(conn *network.GameConn) {
  for {
    req, err := network.Recv(conn);
    if err != nil {
      log.Print(err)
      return
    }

    log.Printf("Client sent:%s\n", req);

    resp := map[string]interface{}{"success":true}

    err = network.Send(conn, resp);
    if err != nil {
      log.Print(err)
      return
    }
  }
}

func main() {
  log.SetFlags(log.Lshortfile)

  ln, err := network.Listen("")
  if err != nil {
    log.Fatal(err)
  }

  for {
    conn, err := network.Accept(ln)
    if err != nil {
      log.Print(err)
    } else {
      go clientConnThread(conn)
    }
  }

  err = ln.Close()
  if err != nil {
    log.Fatal(err)
  }
}
