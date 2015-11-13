package main

import (
  "log"
  "encoding/json"
  "io/ioutil"
  "path/filepath"
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
}

func initWorld()(*World, error){
  var world World
  var region Region

  mapFiles, err := ioutil.ReadDir("./regions/")
  if err != nil {
    log.Printf("Could not read maps dir\n")
    return nil, err
  }

  world.Pcs = make(map[string]*Actor);
  world.Regions = make(map[string]*Region);

  for _, f := range mapFiles {
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

    world.Regions[region.Name] = &region;
  }
  return &world, nil;
}
