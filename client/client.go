package main

import (
  "log"
  "fmt"
  "github.com/nparry0/network"
  "os"
  ui "github.com/gizak/termui"
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

  // Kick off the go routine to handle getting room updates
  go recvUpdates(conn);
  
  // Set up the UI
  startUI();
}

func recvUpdates(conn *network.GameConn) {
  for {
    resp, msgType, err := network.Recv(conn);
    if err != nil {
      log.Fatal(err)
    }

    if msgType == network.TypeRoomUpdate {
      fmt.Printf("Room update! %v\n", resp.RoomUpdate)
    } else if msgType == network.TypeResp && !resp.Resp.Success {
      fmt.Printf("Error, server says: %s\n", resp.Resp.Message)
      os.Exit(1);
    }
  }
}

func startUI() {
  err := ui.Init()
  if err != nil {
    panic(err)
  }
  defer ui.Close()

  //ui.UseTheme("helloworld")
  height := ui.TermHeight()

  descPar := ui.NewPar("This is a room description")
  descPar.Height = (height-3)/2
  //descPar.Width = 50
  descPar.TextFgColor = ui.ColorWhite
  descPar.BorderLabel = "Description"
  descPar.BorderFg = ui.ColorCyan

  strs := []string{
    "You",
    "NPC 1",
    "NPC 2",
  }

  entityList := ui.NewList()
  entityList.Items = strs
  entityList.ItemFgColor = ui.ColorYellow
  entityList.BorderLabel = "List"
  entityList.Height = (height-3)/2
  //entityList.Width = 25
  entityList.Y = 0

  activityPar := ui.NewPar("Something happened\nSomething else happened")
  activityPar.Height = (height-3)/2
  //activityPar.Width = 50
  activityPar.TextFgColor = ui.ColorWhite
  activityPar.BorderLabel = "Activity"
  activityPar.BorderFg = ui.ColorCyan

  cmdPar := ui.NewPar("")
  cmdPar.Height = 3
  //cmdPar.Width = 50
  cmdPar.TextFgColor = ui.ColorWhite
  cmdPar.BorderLabel = "Enter Command"
  cmdPar.BorderFg = ui.ColorCyan


  // build
  ui.Body.AddRows(
    ui.NewRow(
      ui.NewCol(4, 0, entityList),
      ui.NewCol(8, 0, descPar)),
    ui.NewRow(
      ui.NewCol(12, 0, activityPar)),
    ui.NewRow(
      ui.NewCol(12, 0, cmdPar)))

  ui.Body.Align()

  ui.Render(ui.Body)

  ui.Handle("/sys/kbd/C-c", func(ui.Event) {
    ui.StopLoop()
  })
  ui.Handle("/sys/kbd/<space>", func(e ui.Event) {
    cmdPar.Text = cmdPar.Text + " "
    ui.Render(ui.Body)
  })
  ui.Handle("/sys/kbd/<enter>", func(e ui.Event) {
    cmdPar.Text = ""
    ui.Render(ui.Body)
  })
  ui.Handle("/sys/kbd/C-8", func(e ui.Event) {
    cmdPar.Text = cmdPar.Text[:len(cmdPar.Text)-1]
    ui.Render(ui.Body)
  })
  ui.Handle("/sys/kbd", func(e ui.Event) {
    // handle all other key pressing
    k := e.Data.(ui.EvtKbd)
    cmdPar.Text = cmdPar.Text + k.KeyStr  
/*
    switch k.KeyStr{
      case "<space>":
        cmdPar.Text = cmdPar.Text + " "
      default:
        cmdPar.Text = cmdPar.Text + k.KeyStr  
    }
*/

    ui.Render(ui.Body)
  })
  ui.Handle("/sys/wnd/resize", func(e ui.Event) {
    height := ui.TermHeight()
    descPar.Height = (height-3)/2
    entityList.Height = (height-3)/2
    activityPar.Height = (height-3)/2

    ui.Body.Width = ui.TermWidth()
    ui.Body.Align()
    ui.Render(ui.Body)
  })

  ui.Loop()

  /*
  fmt.Printf("%s\n%s\n\nDirections: ", resp.RoomUpdate.Name, resp.RoomUpdate.Desc)
  if resp.RoomUpdate.North {
    fmt.Printf("North ");
  }
  if resp.RoomUpdate.East {
    fmt.Printf("East ");
  }
  if resp.RoomUpdate.South {
    fmt.Printf("South ");
  }
  if resp.RoomUpdate.West {
    fmt.Printf("West");
  }
  for _, pc := range resp.RoomUpdate.Pcs {
    fmt.Printf("%s\n", pc);
  }
  for _, npc := range resp.RoomUpdate.Npcs {
    fmt.Printf("%s\n", npc);
  }
  */
}
