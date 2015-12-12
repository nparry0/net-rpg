package main

import (
  "log"
  "encoding/json"
  "io/ioutil"
  "path/filepath"
  "github.com/nparry0/network"
)

/***** Room *****/
type RoomCoords struct {
  Region string
  Row int
  Col int
}

type Room struct {
  Name string
  Desc string
  Pcs map[string]*Actor
  Npcs map[string]*Actor
  //TODO: Items
  North bool
  South bool
  East bool
  West bool

  // Send commands to the room asynchronously (with a pipe)
  CmdChanReadAsync <-chan network.GameMsg
  CmdChanWriteAsync chan<- network.GameMsg

  // Send commands to the room synchronously (just normal channel)
  CmdChanWriteSync chan RoomHandlerCmd
  CmdChanReadSync chan bool

  UpdateChans map[string]chan<- network.GameMsg
}

// Similar to CmdReq in the GameMsg, but with more info
// Sent in the synchronous chan, but GameMsgs are sent in the async
type RoomHandlerCmd struct {
  Actor *Actor
  Cmd string
  Arg1 string
  UpdateChan chan<- network.GameMsg
}

func (room *Room)createRoomUpdate(msg string)(network.GameMsg){
  pcs := make([]string, len(room.Pcs))

  i := 0
  for k := range room.Pcs{
    pcs[i] = k
    i++
  }

  update := network.GameMsg{RoomUpdate:&network.RoomUpdate{
    Name:room.Name,
    Desc:room.Desc,
    Pcs:pcs,
    Npcs:[]string{},
    North:room.North,
    South:room.South,
    East:room.East,
    West:room.West,
    Message:msg}}
  return update;
}

func roomHandler(room *Room)(){
  log.Printf("Starting handler for room %s\n", room.Name)

  for {
    msg := ""
    select {
      // Sync commands.  Things like adding/removing from a room
      case cmd := <-room.CmdChanWriteSync:
        ret := true
        switch cmd.Cmd {
          case "add":
            log.Printf("Received an add command from %s\n", cmd.Actor.Name)
            room.Pcs[cmd.Actor.Name] = cmd.Actor
            room.UpdateChans[cmd.Actor.Name] = cmd.UpdateChan
          default:
            log.Printf("Invalid sync game message command %s\n", cmd.Cmd)
            ret = false
        }
        
        room.CmdChanReadSync <- ret

      // Async commands.  Basically everything that occurs within the room and doesn't require immediate feedback.
      case cmd := <-room.CmdChanReadAsync:
        switch cmd.CmdReq.Cmd {
          case "say":
            msg = cmd.CmdReq.Actor + " says \"" + cmd.CmdReq.Arg1 + "\""
          default:
            log.Printf("Invalid async game message command\n")
        }
    }

    // Something happened.  Create a room update and send it to everyone in the room
    update := room.createRoomUpdate(msg);
    for _, u := range room.UpdateChans {
      u <- update;
    }
  }
}



/***** Region *****/
type Region struct {
  Name string
  Rows [][]*Room
} 


/***** World *****/
const (
  NoDirection int = iota
  North 
  South 
  East
  West
  Northeast
  Northwest
  Southeast
  Southwest
  Up
  Down
)

type RoomFetcherMsg struct {
  Direction int  
  RoomCoords *RoomCoords
  Room *Room
}

type World struct {
  Pcs map[string]*Actor // PCs are globally unique
  Regions map[string]*Region
  RoomFetcherInChan chan RoomFetcherMsg // Ask for info about rooms
  RoomFetcherOutChan chan RoomFetcherMsg // Ask for info about rooms
} 

// Fetches channels for a room given where you are and where you want to go.
// This allows a client conn to ask to change rooms
func (world World) roomFetcher() {
    var msg RoomFetcherMsg

    log.Printf("Started room fetcher\n")

    for {
      msg = <-world.RoomFetcherInChan
      // TODO: tons more validation in this switch
      switch msg.Direction {
      case NoDirection:
        msg.Room = world.Regions[msg.RoomCoords.Region].Rows[msg.RoomCoords.Row][msg.RoomCoords.Row]
      case North:
        msg.RoomCoords.Row -= 1;
        msg.Room = world.Regions[msg.RoomCoords.Region].Rows[msg.RoomCoords.Row][msg.RoomCoords.Row]
      case South:
        msg.RoomCoords.Row += 1;
        msg.Room = world.Regions[msg.RoomCoords.Region].Rows[msg.RoomCoords.Row][msg.RoomCoords.Row]
      case East:
        msg.RoomCoords.Col += 1;
        msg.Room = world.Regions[msg.RoomCoords.Region].Rows[msg.RoomCoords.Row][msg.RoomCoords.Row]
      case West:
        msg.RoomCoords.Col -= 1;
        msg.Room = world.Regions[msg.RoomCoords.Region].Rows[msg.RoomCoords.Row][msg.RoomCoords.Row]
      default:
        log.Printf("Unrecognized direction (%d)\n", msg.Direction)
        msg.Room = nil
        msg.RoomCoords = nil
      }
      world.RoomFetcherOutChan <- msg
    }
}

func NewWorld()(*World, error){
  var world World
  var region Region

  regionFiles, err := ioutil.ReadDir("./regions/")
  if err != nil {
    log.Printf("Could not read maps dir\n")
    return nil, err
  }

  world.Pcs = make(map[string]*Actor);
  world.Regions = make(map[string]*Region);

  // Load every region file in the regions folder
  for _, f := range regionFiles{
    if filepath.Ext(f.Name()) != ".json" {
      continue;
    }

    log.Printf("Loading region %s\n", f.Name())

    file, err := ioutil.ReadFile("regions/" + f.Name())
    if err != nil {
      log.Printf("Could not read regions/%s\n", f.Name())
      return nil, err
    }

    err = json.Unmarshal(file, &region)
    if err != nil {
      log.Printf("Region %s exists, but config file is corrupt\n", f.Name())
      return nil, err
    }

    // Create a room handler goroutine and pipes to talk to and from each room
    for _, h := range region.Rows {
      for _, room := range h {
        //log.Printf(room);
        room.Pcs = make(map[string]*Actor);
        room.CmdChanWriteAsync, room.CmdChanReadAsync = network.NewPipe()
        room.CmdChanWriteSync = make(chan RoomHandlerCmd)
        room.CmdChanReadSync = make(chan bool)
        room.UpdateChans = make(map[string]chan<- network.GameMsg);
        go roomHandler(room)
      }
    }
    
    world.Regions[region.Name] = &region;
    world.RoomFetcherInChan = make(chan RoomFetcherMsg, 1)
    world.RoomFetcherOutChan = make(chan RoomFetcherMsg, 1)
    go world.roomFetcher()
  }

  return &world, nil;
}
