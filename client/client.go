package main

import (
  "log"
  "fmt"
  "github.com/nparry0/network"
  "os"
)

func main() {
  log.SetFlags(log.Lshortfile)

  // Prompt the user for their name and pass
  var name, pass, character string
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
  resp, msgType, err := network.Recv(conn);
  if err != nil {
    log.Fatal(err)
  }

  // Did we log in successfully?
  if msgType == network.TypeResp && resp.Resp.Success == true {
    fmt.Printf("Login Successful :)\n");
  } else {
    fmt.Printf("Login Failed :(\n");
    fmt.Printf("Server says: %s\n", resp.Resp.Message)
    os.Exit(1);
  }

  // Character select
  fmt.Printf("Available characters:\n")
  for index, each := range resp.Resp.Data{
    fmt.Printf("%d)  %s\n", index+1, each)
  }
  fmt.Printf("Which character would you like to play? ")
  fmt.Scanln(&character)

  req = network.GameMsg{AssumeActorReq:&network.AssumeActorReq{Actor:character}}

  // Send the request
  err = network.Send(conn, req);
  if err != nil {
    log.Fatal(err)
  }
  resp, msgType, err = network.Recv(conn);
  if err != nil {
    log.Fatal(err)
  }

  // Did we log in successfully?
  if msgType == network.TypeResp && resp.Resp.Success == true {
    fmt.Printf("Successfully selected character :)\n");
  } else {
    fmt.Printf("Could not select character :(\n");
    fmt.Printf("Server says: %s\n", resp.Resp.Message)
    os.Exit(1);
  }
}
