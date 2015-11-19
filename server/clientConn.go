package main

import (
  "log"
  "github.com/nparry0/network"
  "strconv"
)

type ClientConn struct {
  user *User;  
  actor *Actor;  
  GameConn *network.GameConn // TODO: maybe split the send and recieve parts here, for better exclusive access?

  RoomChan <-chan network.GameMsg       // Read room updates
  CmdChan chan<- network.GameMsg        // Send commands to room

  SendChanRead <-chan network.GameMsg   // Pipe for sending msgs to client 
  SendChanWrite chan<- network.GameMsg
}

var MIN_CLIENT_VERSION int = 1

func (client ClientConn) clientSender() {
  for {
    resp := <-client.SendChanRead
    err := network.Send(client.GameConn, resp)
    if err != nil {
      log.Print(err)
      return
    }
  }
}

/* Each client gets one of these threads to listen to it, send data to it, 
   and handle validation for the life of the client */
func (client ClientConn) clientReceiver() {
  for {
    resp := network.GameMsg{Resp:&network.Resp{Success:false}}

    req, msgType, err := network.Recv(client.GameConn)
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
      } else if client.user != nil {
        log.Printf("Multiple login attempts from %s\n", req.LoginReq.Username)
        resp.Resp.Message = "ERROR: Duplicate login attempt."
        break;
      }

      client.user, err = userLogin(req.LoginReq.Username, req.LoginReq.Password)
      if err != nil {
        log.Printf("Failed login for user %s\n", req.LoginReq.Username)
        resp.Resp.Message = "Incorrect username or password.  Please try again."
        break;
      } 

      // Successfully logged in
      log.Printf("Successful login from %s\n", req.LoginReq.Username)
      resp.Resp.Message = "Successfully logged in as " + client.user.Username
      resp.Resp.Data = client.user.Actors
      resp.Resp.Success = true;

    // Request to assume actor
    case network.TypeAssumeActorReq:
      if client.user == nil {
        log.Printf("Attempted to assume %s without logging in\n", req.AssumeActorReq.Actor)
        resp.Resp.Message = "You must log in first"
        break;
      } else if client.actor != nil {
        //TODO: Allow user to switch players
        log.Printf("User %s attempted to assume a second actor %s\n", client.user.Username, req.AssumeActorReq.Actor)
        resp.Resp.Message = "You cannot play two characters at once."
        break;
      }

      client.actor, err = assumeActor(req.AssumeActorReq.Actor, client.user)
      if err != nil {
        log.Printf("User %s failed to assume actor %s\n", client.user.Username, req.AssumeActorReq.Actor)
        resp.Resp.Message = "Unable to use that character"
        break;
      } 

      // Successfully assumed an actor.  Set up a pipe with the room, add the new actor to the global list,
      // drop that actor in a room, and send the client a success message along with their first room update
      gWorld.RoomFetcherInChan <- RoomFetcherMsg{Direction:NoDirection, RoomCoords:&client.actor.Coords}
      fetcherMsg := <- gWorld.RoomFetcherOutChan;

      log.Printf("DEBUG: client conn fetcher msg: %v\n", fetcherMsg)

      log.Printf("User %s successfully assumed actor %s\n", client.user.Username, req.AssumeActorReq.Actor)
      resp.Resp.Message = "Successfully assumed character " + client.actor.Name
      resp.Resp.Success = true;

    default:
      log.Printf("Recv'd invalid type: %d\n", msgType)
      resp.Resp.Message = "ERROR: Invalid message type (" + strconv.Itoa(msgType) + ")."
    }

    // Send response and listen for another message
    client.SendChanWrite <- resp
  }
}

func NewClientConn(gameConn *network.GameConn) (*ClientConn) {
  var client ClientConn
  client.GameConn = gameConn
	client.SendChanWrite, client.SendChanRead = network.NewPipe()
  go client.clientReceiver()
  go client.clientSender()
  return &client;
}
