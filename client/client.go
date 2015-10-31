package main

import (
  "log"
  "fmt"
  "github.com/nparry0/network"
)

func main() {
    log.SetFlags(log.Lshortfile)

    // Prompt the user for their name and pass
    var name, pass string
    fmt.Printf("*** Stuff N' Things the RPG ***\n")
    fmt.Printf("Login:")
    fmt.Scanln(&name)
    fmt.Printf("Password:")
    fmt.Scanln(&pass)
  
    fmt.Printf("Logging in as %s\n", name)
    
    req := network.GameMsg{LoginReq:&network.LoginReq{Version:1, Username:name, Password:pass}}
    //log.Printf("main: %s\n", req);

    // Connect to the server
    conn, err := network.Connect("")
    if err != nil {
      log.Fatal(err)
    }

    // Send the request
    err = network.Send(conn, req);
    if err != nil {
      log.Fatal(err)
    }

    // Get all the info for our user back.
    resp, err := network.Recv(conn);
    if err != nil {
      log.Fatal(err)
    }

    fmt.Printf("Response was %s \n", resp);
}
