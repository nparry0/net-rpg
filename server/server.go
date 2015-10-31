package main

import (
  "log"
  "github.com/nparry0/network"
  "strconv"
)

var MIN_CLIENT_VERSION int = 1

/* Each client gets one of these threads to listen to it, send data to it, 
   and handle validation for the life of the client */
func clientConnThread(conn *network.GameConn) {
  var user *User;  
  var actor *Actor;  

  for {
    resp := network.GameMsg{Resp:&network.Resp{Success:false}}

    req, msgType, err := network.Recv(conn)
    if err != nil {
      log.Print(err)
      return
    }

    switch msgType {

    // Log in request
    case network.TypeLoginReq:
      // A few sanity checks
      if req.LoginReq.Version < MIN_CLIENT_VERSION {
        log.Printf("Attempt to log in with old client version %d (min version is %d)\n", req.LoginReq.Version, MIN_CLIENT_VERSION)
        resp.Resp.Message = "Your client version is too old for this server.  Please update to the latest version."
        break;
      }
      if user != nil {
        log.Printf("Multiple login attempts from %s\n", req.LoginReq.Username)
        resp.Resp.Message = "ERROR: Duplicate login attempt."
        break;
      }

      user, err = userLogin(req.LoginReq.Username, req.LoginReq.Password)
      if err != nil {
        log.Printf("Failed login for user %s\n", req.LoginReq.Username)
        resp.Resp.Message = "Incorrect username or password.  Please try again."
        break;
      } 

      // Successfully logged in
      log.Printf("Successful login from %s\n", req.LoginReq.Username)
      resp.Resp.Message = "Successfully logged in as " + user.Username
      resp.Resp.Data = user.Actors
      resp.Resp.Success = true;

    // Request to assume actor
    case network.TypeAssumeActorReq:
      if user == nil {
        log.Printf("Attempted to assume %s without logging in\n", req.AssumeActorReq.Actor)
        resp.Resp.Message = "You must log in first"
        break;
      }

      actor, err = assumeActor(req.AssumeActorReq.Actor, user)
      if err != nil {
        log.Printf("User %s failed to assume actor %s\n", user.Username, req.AssumeActorReq.Actor)
        resp.Resp.Message = "Unable to use that character"
        break;
      } 

      // Successfully logged in
      log.Printf("User %s successfully assumed actor %s\n", user.Username, req.AssumeActorReq.Actor)
      resp.Resp.Message = "Successfully assumed character " + actor.Name
      resp.Resp.Success = true;

    default:
      log.Printf("Recv'd invalid type: %d\n", msgType)
      resp.Resp.Message = "ERROR: Invalid message type (" + strconv.Itoa(msgType) + ")."
    }

    // Send response and listen for another message
    err = network.Send(conn, resp)
    if err != nil {
      log.Print(err)
      return
    }
  }
}

func main() {
  log.SetFlags(log.Lshortfile)

  ln, err := network.Listen("")
  if err != nil {
    log.Fatal(err)
  }

  for {
    conn, err := network.Accept(ln)
    if err != nil {
      log.Print(err)
    } else {
      go clientConnThread(conn)
    }
  }

  err = ln.Close()
  if err != nil {
    log.Fatal(err)
  }
}
