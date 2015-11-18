package main

import (
  "log"
  "encoding/json"
  "io/ioutil"
  "path/filepath"
  "github.com/nparry0/network"
)

// TODO: Make some of these (and other) structs' members private
type World struct {
  Pcs map[string]*Actor // PCs are globally unique
  Regions map[string]*Region
} 

type Region struct {
  Name string
  Rows [][]*Room
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

  // Send commands to the room
  CmdChanRead <-chan network.GameMsg
  CmdChanWrite chan<- network.GameMsg

  // Get updates from the room
  UpdateChanRead <-chan network.GameMsg
  UpdateChanWrite chan<- network.GameMsg
}

func roomHandler(room *Room)(){
    log.Printf("Starting handler for room %s\n", room.Name)
}

func initWorld()(*World, error){
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
        room.CmdChanWrite, room.CmdChanRead = network.NewPipe()
        room.UpdateChanWrite, room.UpdateChanRead = network.NewPipe()
        go roomHandler(room)
      }
    }
    
    
    world.Regions[region.Name] = &region;
  }
  return &world, nil;
}
