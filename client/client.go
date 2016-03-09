package main

import (
  "log"
  "fmt"
  "github.com/nparry0/network"
  "os"
  //ui "github.com/gizak/termui"
  ui "gopkg.in/gizak/termui.v1"
  //"strconv"
  "strings"
  "errors"
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

  update := make(chan *network.RoomUpdate)
  activity:= make(chan string)

	sendChanWrite, sendChanRead := network.NewPipe()

  go sendCmds(conn, sendChanRead)
  go recvUpdates(conn, update, activity)
  
  // Set up the UI
  // TODO: make this whole deal an object, add character and other stuff to it as a member 
  startUI(update, activity, sendChanWrite, character)
}

func sendCmds(conn *network.GameConn, cmd <-chan network.GameMsg) {
  for {
    msg := <-cmd
    err := network.Send(conn, msg)
    if err != nil {
      log.Print(err)
      return
    }
  }
}

func recvUpdates(conn *network.GameConn, update chan *network.RoomUpdate, activity chan string) {
  for {
    resp, msgType, err := network.Recv(conn);
    if err != nil {
      log.Fatal(err)
    }

    if msgType == network.TypeRoomUpdate {
      update <- resp.RoomUpdate
    } else if msgType == network.TypeResp {
      if resp.Resp.Message != "" {
        activity <- resp.Resp.Message
      } else if !resp.Resp.Success {
        activity <- "Invalid command (type 'help' and hit enter for assistance)"
      }
    }
  }
}

func startUI(update chan *network.RoomUpdate, activity chan string, cmd chan<- network.GameMsg, character string) {
  err := ui.Init()
  if err != nil {
    panic(err)
  }
  defer ui.Close()

  //ui.UseTheme("helloworld")
  height := ui.TermHeight()

  descPar := ui.NewPar("")
  descPar.Height = ((height-3)/2) - 5
  //descPar.Width = 50
  descPar.TextFgColor = ui.ColorWhite
  descPar.Border.Label = ""
  descPar.Border.FgColor = ui.ColorCyan

  exitPar := ui.NewPar("")
  exitPar.Height = 5
  exitPar.TextFgColor = ui.ColorWhite
  exitPar.Border.Label = "Exits"
  exitPar.Border.FgColor = ui.ColorCyan



  entityList := ui.NewList()
  entityList.ItemFgColor = ui.ColorYellow
  entityList.Border.Label = "Entities"
  entityList.Height = (height-3)/2
  entityList.Y = 0

  activityList := ui.NewList()
  activityList.ItemFgColor = ui.ColorWhite
  activityList.Border.Label = "Activity"
  activityList.Height = (height-3)/2
  activityList.Y = 0

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
      ui.NewCol(8, 0, descPar, exitPar)),
    ui.NewRow(
      ui.NewCol(12, 0, activityList)),
    ui.NewRow(
      ui.NewCol(12, 0, cmdPar)))

  ui.Body.Align()

  ui.Render(ui.Body)

  evt := ui.EventCh()
  redraw := make(chan bool)
  done := make(chan bool)

  clearActivity := false
  var prevCmds []string
  prevCmd := 0

  for {
    select {
    // Get an event from the keyboard, mouse, screen resize, etc
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
            msg, err := parseCmd(cmdPar.Text)
            if err != nil {
              //TODO: alert the user somehow
              break
            }
            // Set a flag so we know to clear the activity list on the next update
            if msg.CmdReq.Cmd == "go" {
              // go command will produce a room update or an error msg
              // both with set this back to false, the room update after clearing the list
              clearActivity = true 
            }
            cmd <- msg
            prevCmds = append(prevCmds, cmdPar.Text)
            prevCmd = len(prevCmds)
            cmdPar.Text = ""
            go func() { redraw <- true }()
          case ui.KeyArrowUp:
            if prevCmd > 0 {
              prevCmd--;
              cmdPar.Text = prevCmds[prevCmd]
              go func() { redraw <- true }()
            }
          case ui.KeyArrowDown:
            if prevCmd < (len(prevCmds)-1) {
              prevCmd++;
              cmdPar.Text = prevCmds[prevCmd]
              go func() { redraw <- true }()
            } else if prevCmd == (len(prevCmds)-1) {
              prevCmd++;
              cmdPar.Text = "" 
              go func() { redraw <- true }()
            }
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
        descPar.Height = ((height-3)/2) - 5
        entityList.Height = (height-3)/2
        activityList.Height = (height-3)/2

        if len(activityList.Items) > (activityList.Height-2) {
          //Slice the top off
          activityList.Items = activityList.Items[len(activityList.Items)-(activityList.Height-2):len(activityList.Items)]
        }

        ui.Body.Width = ui.TermWidth()
        ui.Body.Align()
        go func() { redraw <- true }()
      }
    // We got a room update from the server
    case u := <-update:
      var entities []string
      for _,pc := range u.Pcs {
        entities = append(entities, pc + " (PC)")
      }
      entities = append(entities, u.Npcs...)
      entityList.Items = entities;

      if clearActivity {
        activityList.Items = nil
        clearActivity = false
      }
      if u.Message != "" {
        addActivity(activityList, u.Message)
      }

      exitStr := ""
      if u.North {
        exitStr += "       ^ North\n"
      } else {
        exitStr += "              \n"
      }
      if u.West{
        exitStr += "West < "
      } else {
        exitStr += "       "
      }
      if u.East {
        exitStr += "  > East \n"
      } else {
        exitStr += "         \n"
      }
      if u.South {
        exitStr += "       v South"
      }

      exitPar.Text =  exitStr
      descPar.Text = u.Desc
      descPar.Border.Label = u.Name

      go func() { redraw <- true }()

    // Local write request to activity dialog
    case a := <-activity:
      clearActivity = false 
      addActivity(activityList, a)
      go func() { redraw <- true }()
    // Request to redraw
    case <-redraw:
      ui.Render(ui.Body)
    // We are done
    case <-done:
      return
    }
  }
}

func addActivity(list *ui.List, item string) {
  if list != nil {
    list.Items = append(list.Items, item)
    if len(list.Items) > (list.Height-2) {
      //Slice the top one off
      list.Items = list.Items[1:len(list.Items)]
    }
  }
}

func parseCmd(cmd string)(network.GameMsg, error) {
    msg := network.GameMsg{CmdReq:&network.CmdReq{Cmd:""}}

    words := strings.Fields(cmd)
    if len(words) == 0 {
      return msg, errors.New("Please type a command")
    }
    msg.CmdReq.Cmd = words[0]
    msg.CmdReq.Arg1 = strings.Join(words[1:], " ")
    return msg, nil
}
