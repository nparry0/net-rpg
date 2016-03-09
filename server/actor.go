package main

import (
  "log"
  "encoding/json"
  "io/ioutil"
  "errors"
)

type Actor struct {
  Name string
  Hp int
  MaxHp int
}

func (actor *Actor) DecHp(hp int)(int) {
  actor.Hp -= hp;

  if actor.Hp < 0 {
    actor.Hp = 0
  } else if actor.Hp > actor.MaxHp {
    actor.Hp = actor.MaxHp
  }
  return actor.Hp
}

type Pc struct {
  Actor
}

type Npc struct {
  Actor
  Aggro int
  Mobile bool
  Respawn int
}

type PcJson struct {
  Pc Pc
  Coords RoomCoords
}

//TODO: When AI is smart enough to handle more aggro levels, we could have more levels above and below hostile
// The NPCs move through different aggro levels when they are attacked by a PC
// These are values in a json file, don't change them
const (
  AggroNone = 0     // Does not attack PC
  AggroHostile = 10 // Attacks PCs on sight
)

func NewPc(actorName string, user *User)(*Pc, *RoomCoords, error) {
  var pcJson PcJson

  // Check for a sanitary name
  if !checkName(actorName) {
    log.Printf("Actor %s is not a sanitary actor name\n", actorName)
    return nil, nil, errors.New("Not a sanitary actor name")
  }

  // Read the json file for the actor after verifying that it actually belongs to this user
  for _, a := range user.Actors {
    if a == actorName {
      file, err := ioutil.ReadFile("pcs/" + actorName + ".json")
      if err != nil {
        log.Printf("Actor %s does not exist, although it is listed as valid for user %s\n", actorName, user.Username)
        return nil, nil, err
      }

      err = json.Unmarshal(file, &pcJson)
      if err != nil {
        log.Printf("Actor %s exists, but config file is corrupt\n", actorName)
        return nil, nil, err
      }
      return &pcJson.Pc, &pcJson.Coords, nil
    }
  }

  log.Printf("Actor %s does not exist for user %s\n", actorName, user.Username)
  return nil, nil, errors.New("Actor does not exist")
}
