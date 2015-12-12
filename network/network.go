package network

import (
//  "log"
  "net"
  "encoding/json"
  "crypto/tls"
  "crypto/rand"
  "errors"
)

/*********************/
/***** Comms API *****/
/*********************/
const (
  TypeError int = iota
  TypeResp
  TypeLoginReq
  TypeAssumeActorReq
  TypeCmdReq
  TypeRoomUpdate
)

type Resp struct {
  Success bool
  Message string
  Data []string
}

type LoginReq struct {
  Version int //Not every struct needs a version field, but login does for client version
  Username string
  Password string
}

type AssumeActorReq struct {
  Actor string
}

type CmdReq struct {
  Actor string
  Cmd string
  Arg1 string
}

type RoomUpdate struct {
  Name string
  Desc string
  Pcs []string
  Npcs []string
  North bool
  South bool
  East bool
  West bool
  Message string
}

type GameMsg struct {
  Resp *Resp
  LoginReq *LoginReq
  AssumeActorReq *AssumeActorReq
  CmdReq *CmdReq
  RoomUpdate *RoomUpdate
}

/*************************/
/***** Net functions *****/
/*************************/
type GameConn struct {
  conn net.Conn
  dec * json.Decoder
  enc * json.Encoder
}

func GetGameMsgType(msg *GameMsg)(int, error){
  var respType int

  if msg.Resp != nil {
    respType = TypeResp
  } else if msg.LoginReq != nil {
    respType = TypeLoginReq
  } else if msg.AssumeActorReq != nil {
    respType = TypeAssumeActorReq
  } else if msg.CmdReq != nil {
    respType = TypeCmdReq
  } else if msg.RoomUpdate != nil {
    respType = TypeRoomUpdate
  } else {
    return TypeError, errors.New("Invalid Game Message")
  }
  return respType, nil
}

func Listen(port string)(net.Listener, error){
  if port == "" {
    port = ":10101"
  }

  //TODO: What will the real file paths be?
  cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
  if err != nil {
    return nil, err
  }

  config := tls.Config{Certificates: []tls.Certificate{cert} }
  config.Rand = rand.Reader

  ln, err := tls.Listen("tcp", ":10101", &config)
  if err != nil {
    return nil, err
  }

  return ln, nil
}

func Accept(ln net.Listener)(*GameConn, error){
  var conn GameConn;

  c, err := ln.Accept()
  if err != nil {
    return nil, err 
  }

  conn.conn = c
  conn.enc = json.NewEncoder(conn.conn)
  conn.dec = json.NewDecoder(conn.conn)
  return &conn, nil
}

func Connect(server string)(*GameConn, error){
  var conn GameConn;

  //TODO: Name Resolution
  if server == "" {
    server = "127.0.0.1:10101"
  }

  //TODO: Actual hostname and a real cert
  c, err := tls.Dial("tcp", server, &tls.Config{ServerName:"localhost", InsecureSkipVerify:true})
  if err != nil {
    return nil, err
  }

  conn.conn = c
  conn.enc = json.NewEncoder(conn.conn)
  conn.dec = json.NewDecoder(conn.conn)
  return &conn, nil
}

func Send(conn *GameConn, req interface{})(error) {
  //log.Printf("Send: %s\n", req);
  err := conn.enc.Encode(req)
  return err
}

func Recv(conn *GameConn)(*GameMsg, int, error) {
  var resp GameMsg

  err := conn.dec.Decode(&resp)
  if err != nil {
    return nil, TypeError, err
  }

  respType, err := GetGameMsgType(&resp)
  if err != nil {
    return nil, TypeError, err
  }

  return &resp, respType, nil;
}
