package main

import (
  "log"
  "encoding/json"
  "io/ioutil"
  "path/filepath"
  "github.com/nparry0/network"
  "math/rand"
  "time"
  "strconv"
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
  Pcs map[string]*Pc
  Npcs map[string]*Npc
  //TODO: Items
  North bool
  South bool
  East bool
  West bool

  // JSON entities
  NpcEntities []*Npc

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
  Pc *Pc
  Cmd string
  Arg1 string
  UpdateChan chan<- network.GameMsg
}

func (room *Room)createRoomUpdate(msg string)(network.GameMsg){
  pcs := make([]string, len(room.Pcs))
  npcs := make([]string, len(room.Npcs))

  i := 0
  for k := range room.Pcs {
    pcs[i] = k
    i++
  }
  i = 0
  for k := range room.Npcs {
    npcs[i] = k
    i++
  }

  update := network.GameMsg{RoomUpdate:&network.RoomUpdate{
    Name:room.Name,
    Desc:room.Desc,
    Pcs:pcs,
    Npcs:npcs,
    North:room.North,
    South:room.South,
    East:room.East,
    West:room.West,
    Message:msg}}
  return update;
}

func roomHandler(room *Room)(){
  log.Printf("Starting handler for room %s\n", room.Name)

  rand := rand.New(rand.NewSource(time.Now().UnixNano()))

  for {
    msg := ""
    select {
      // Sync commands.  Things like adding/removing from a room
      case cmd := <-room.CmdChanWriteSync:
        ret := true
        switch cmd.Cmd {
          case "add":
            room.Pcs[cmd.Pc.Name] = cmd.Pc
            room.UpdateChans[cmd.Pc.Name] = cmd.UpdateChan
          case "rem":
            delete(room.Pcs, cmd.Pc.Name)
            delete(room.UpdateChans, cmd.Pc.Name)
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
          case "attack":
            // TODO: Attack PCs if the PC has PvP on
            if target, ok := room.Npcs[cmd.CmdReq.Arg1]; ok {
              dmg := rand.Intn(10)+1; // 1-10
              msg = cmd.CmdReq.Actor + " attacks " + cmd.CmdReq.Arg1 + " for " + strconv.Itoa(dmg) + " damage."
              if target.DecHp(dmg) <= 0 {
                msg += " " + cmd.CmdReq.Arg1 + " has died.\n"
                delete(room.Npcs, cmd.CmdReq.Arg1)
              }
            }
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
      var curRoom *Room
      msg = <-world.RoomFetcherInChan

      // Check that the room they are in now actually exists
      region := world.Regions[msg.RoomCoords.Region]
      if region == nil || 
         msg.RoomCoords.Row < 0 || 
         msg.RoomCoords.Row > len(region.Rows) || 
         msg.RoomCoords.Col < 0 || 
         msg.RoomCoords.Col > len(region.Rows[msg.RoomCoords.Row]) {
        msg.Room = nil
        msg.RoomCoords = nil
        world.RoomFetcherOutChan <- msg
        continue
      }

      curRoom = region.Rows[msg.RoomCoords.Row][msg.RoomCoords.Col]

      // Check that the room we are going to exists 
      switch msg.Direction {
      case NoDirection:
        msg.Room = region.Rows[msg.RoomCoords.Row][msg.RoomCoords.Col]
      case North:
        if curRoom.North {
          msg.RoomCoords.Row -= 1;
          if  msg.RoomCoords.Row < 0 {
            msg.RoomCoords.Row = 0;
          }
        }
        msg.Room = region.Rows[msg.RoomCoords.Row][msg.RoomCoords.Col]
      case South:
        if curRoom.South {
          msg.RoomCoords.Row += 1;
          len := len(region.Rows)
          if  msg.RoomCoords.Row >= len {
            msg.RoomCoords.Row = len-1;
          }
        }
        msg.Room = region.Rows[msg.RoomCoords.Row][msg.RoomCoords.Col]
      case East:
        if curRoom.East {
          msg.RoomCoords.Col += 1;
          len := len(region.Rows[msg.RoomCoords.Row])
          if  msg.RoomCoords.Col >= len {
            msg.RoomCoords.Col = len-1;
          }
        }
        msg.Room = region.Rows[msg.RoomCoords.Row][msg.RoomCoords.Col]
      case West:
        if curRoom.West {
          msg.RoomCoords.Col -= 1;
          if  msg.RoomCoords.Col < 0 {
            msg.RoomCoords.Col = 0;
          }
        }
        msg.Room = region.Rows[msg.RoomCoords.Row][msg.RoomCoords.Col]
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
        room.Pcs = make(map[string]*Pc);
        room.Npcs = make(map[string]*Npc);
        room.CmdChanWriteAsync, room.CmdChanReadAsync = network.NewPipe()
        room.CmdChanWriteSync = make(chan RoomHandlerCmd)
        room.CmdChanReadSync = make(chan bool)
        room.UpdateChans = make(map[string]chan<- network.GameMsg);
        for _, npc := range room.NpcEntities {
          room.Npcs[npc.Name] = npc 
        }
    
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
