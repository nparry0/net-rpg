package network

import (
    "net"
    "encoding/json"
    "crypto/tls"
)

type GameConn struct {
  conn net.Conn
  dec * json.Decoder
  enc * json.Encoder
}

func Connect(server string)(*GameConn, error){
    var conn GameConn;

    //TODO: Name Resolution
    if server == "" {
      server = ":10101"
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

func Send(conn *GameConn, req map[string]interface{})(error) {
    err := conn.enc.Encode(req)
    return err
}

func Recv(conn *GameConn)(map[string]interface{}, error) {
    var resp map[string]interface{} 

    err := conn.dec.Decode(&resp)
    if err != nil {
      return nil, err
    }
    return resp, nil;
}
