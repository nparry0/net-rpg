package main

import (
  "log"
  "fmt"
  "github.com/nparry0/network"
  "os"
  //ui "github.com/gizak/termui"
  ui "gopkg.in/gizak/termui.v1"
  //"strconv"
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
      //fmt.Printf("Room update! %v\n", resp.RoomUpdate)
    } else if msgType == network.TypeResp && !resp.Resp.Success {
      //fmt.Printf("Error, server says: %s\n", resp.Resp.Message)
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
  descPar.Border.Label = "Description"
  descPar.Border.FgColor = ui.ColorCyan

  strs := []string{
    "You",
    "NPC 1",
    "NPC 2",
  }

  entityList := ui.NewList()
  entityList.Items = strs
  entityList.ItemFgColor = ui.ColorYellow
  entityList.Border.Label = "List"
  entityList.Height = (height-3)/2
  //entityList.Width = 25
  entityList.Y = 0

  activityPar := ui.NewPar("Something happened\nSomething else happened")
  activityPar.Height = (height-3)/2
  //activityPar.Width = 50
  activityPar.TextFgColor = ui.ColorWhite
  activityPar.Border.Label = "Activity"
  activityPar.Border.FgColor = ui.ColorCyan

  cmdPar := ui.NewPar("")
  cmdPar.Height = 3
  //cmdPar.Width = 50
  cmdPar.TextFgColor = ui.ColorWhite
  cmdPar.Border.Label = "Enter Command"
  cmdPar.Border.FgColor = ui.ColorCyan


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

  evt := ui.EventCh()
  redraw := make(chan bool)
  done := make(chan bool)

  for {
    select {
    case e := <-evt:
      switch e.Type {
      case ui.EventKey:
        switch e.Ch {
        case 0: // e.Key is valid if e.Ch is 0
          switch e.Key {
          case ui.KeyBackspace2:
            fallthrough
          case ui.KeyBackspace:
            len := len(cmdPar.Text)
            if len > 0 {
              cmdPar.Text = cmdPar.Text[:len-1] 
              go func() { redraw <- true }()
            }
          case ui.KeySpace:
            cmdPar.Text += " "
            go func() { redraw <- true }()
          case ui.KeyEsc:
            fallthrough
          case ui.KeyCtrlC:
            return
          case ui.KeyEnter:
            cmdPar.Text = ""
            go func() { redraw <- true }()
            //TODO: execute command here
          //default:
          //  cmdPar.Text += strconv.Itoa(int(e.Key))
          //  go func() { redraw <- true }()
          }
        default:
          cmdPar.Text += string(e.Ch)
          go func() { redraw <- true }()
        }
      case ui.EventResize:
        height := ui.TermHeight()
        descPar.Height = (height-3)/2
        entityList.Height = (height-3)/2
        activityPar.Height = (height-3)/2

        ui.Body.Width = ui.TermWidth()
        ui.Body.Align()
        go func() { redraw <- true }()
      }
    case <-done:
      return
    case <-redraw:
      ui.Render(ui.Body)
    }
  }

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
