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
  Attack int
  Intelligence int
  Defense int
  Evasion int
  Coords RoomCoords
}

func assumeActor(actorName string, user *User)(*Actor, error) {
  var actor Actor

  // Check for a sanitary name
  if !checkName(actorName) {
    log.Printf("Actor %s is not a sanitary actor name\n", actorName)
    return nil, errors.New("Not a sanitary actor name")
  }

  // Read the json file for the actor after verifying that it actually belongs to this user
  for _, a := range user.Actors {
    if a == actorName {
      file, err := ioutil.ReadFile("pcs/" + actorName + ".json")
      if err != nil {
        log.Printf("Actor %s does not exist, although it is listed as valid for user %s\n", actorName, user.Username)
        return nil, err
      }

      err = json.Unmarshal(file, &actor)
      if err != nil {
        log.Printf("Actor %s exists, but config file is corrupt\n", actorName)
        return nil, err
      }
      return &actor, nil
    }
  }

  log.Printf("Actor %s does not exist for user %s\n", actorName, user.Username)
  return nil, errors.New("Actor does not exist")
}
