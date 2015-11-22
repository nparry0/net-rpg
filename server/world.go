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

  // Get updates from the room
  UpdateChanRead <-chan network.GameMsg
  UpdateChanWrite chan<- network.GameMsg
}

// Similar to CmdReq in the GameMsg, but with more info
type RoomHandlerCmd struct {
  Actor *Actor
  Cmd string
  Args []string
}

func roomHandler(room *Room)(){
  log.Printf("Starting handler for room %s\n", room.Name)

  for {
    ret := true
    cmd := <-room.CmdChanWriteSync

    ret = true
    switch cmd.Cmd {
      case "add":
        log.Printf("Received an add command from %s\n", cmd.Actor.Name)
        room.Pcs[cmd.Actor.Name] = cmd.Actor
      default:
        log.Printf("Invalid game message command\n")
        ret = false
    }
    
    room.CmdChanReadSync <- ret
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
        room.UpdateChanWrite, room.UpdateChanRead = network.NewPipe()
        room.CmdChanWriteSync = make(chan RoomHandlerCmd)
        room.CmdChanReadSync = make(chan bool)
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
