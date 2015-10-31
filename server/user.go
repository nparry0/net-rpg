package main

import (
  "log"
  "encoding/json"
  "io/ioutil"
  "regexp"
  "errors"
)

type User struct {
  Username string
  PasswordHash string
  Email string
  Characters []string
}

func checkName(name string)(bool) {
  //TODO: Maybe init this globally, using MustCompile instead of Compile
  re := regexp.MustCompile("^[a-zA-Z0-9_]*$")
  if re.MatchString(name) {
    return true;
  }
  return false;
}

func userLogin(username string, password string)(*User, error) {
  var user User

  //Check for a sanitary username
  if !checkName(username) {
    log.Printf("User %s is not a sanitary username\n", username)
    return nil, errors.New("Not a sanitary username")
  }

  file, err := ioutil.ReadFile("users/" + username + ".json")
  if err != nil {
    log.Printf("User %s does not exist\n", username)
    return nil, err
  }

  err = json.Unmarshal(file, &user)
  if err != nil {
    log.Printf("User %s exists, but config file is corrupt\n", username)
    return nil, err
  }
  
  return &user, nil
}
