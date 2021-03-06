package main

import (
    // "bufio"
    "log"
    // "strings"
    // "runtime"
    // "sync"
    "net"
    "net/rpc"
    "net/rpc/jsonrpc"
    // "net/http"
    "os"
    "fmt"
    // "errors"
    // "hash/fnv"
    // "time"
    )

func (t *KeySpace)GetPredeccesor(nothing string, reply *Address) error{
    *reply = predecessor
    return nil
}

func (t *KeySpace)GetSuccessor(nothing string, reply *Address) error{
    *reply = successor
    return nil
}

func (t *KeySpace)FindSuccessor(key string, reply *Address) error{
    if successor.to_string()==own_Address.to_string() {
        *reply = successor
    } else {
        if in_between(own_Address.to_string(),key,successor.to_string()) {
            *reply = successor 
        } else {
            conn, err := net.Dial("tcp", successor.Ip + ":" + successor.Port)
            if err != nil {
                fmt.Println("error in FindSuccessor 0", err)
                return err
            }
            client := jsonrpc.NewClient(conn)
            var reply_obj Address
            err = client.Call("KeySpace.FindSuccessor", key, &reply_obj)
            *reply = reply_obj
            if err != nil {
                fmt.Println("error in FindSuccessor 1", err)
                return err
            }
            conn.Close()
        }
    }
    return nil
}


func (t *KeySpace)Notify(addr Address, reply *Address) error{
    prev_predeccesor := predecessor
    if (predecessor.Ip=="" || in_between(predecessor.to_string(),addr.to_string(),own_Address.to_string())) {
        predecessor = addr
    }
    if prev_predeccesor.to_string()!=predecessor.to_string() {
        fmt.Println("predecessor changed to - ",predecessor.to_string())
        if ((!in_between(prev_predeccesor.to_string(),predecessor.to_string(),own_Address.to_string()))){
            for key, value := range store_pred{
                store[key]=value
                delete(store_pred, key)
            }
        }
        if(predecessor.Ip!=""){
            fmt.Println("pred string - ",predecessor.to_string())
            conn, err := net.Dial("tcp", predecessor.Ip+":"+predecessor.Port)
            if err != nil {
                log.Fatal("Connectiong in notify:", err)
            }
            client := jsonrpc.NewClient(conn)
            err = client.Call("KeySpace.GetStore", "", &store_pred)
            if err != nil {
                log.Fatal("error while getting keyval:", err)
            }
            conn.Close()
        }

    } 
    if (successor.to_string()==own_Address.to_string()){
        successor=addr
    }
    return nil
}

func (t *KeySpace)Getkeyval(addr Address, reply *map[string]string) error{
    reply_obj := make(map[string]string)
    for key, value := range store{
        if(hash(key) <= hash(addr.to_string())){
            reply_obj[key]=value
            delete(store, key)
        }
    }
    *reply = reply_obj
    return nil
}

func (t *KeySpace)GetStore(nothing string, reply *map[string]string) error{
    // fmt.Println("yoyo")
    reply_obj := make(map[string]string)
    for key, value := range store{
        reply_obj[key]=value
    }
    *reply = reply_obj
    return nil
}


func Stabilize() error{
    if successor.to_string() == own_Address.to_string() {
        return nil
    }
    prev_successor := successor
    prev_succ_succ := succ_succ
    if(predecessor.Ip!=""){
        conn, err := net.Dial("tcp", predecessor.Ip + ":" + predecessor.Port)
        if err != nil {
            predecessor = Address{"",""}
        } else {
            conn.Close()
        }
    }
    conn, err := net.Dial("tcp", successor.Ip + ":" + successor.Port)
    if err != nil {
        successor=succ_succ
        if successor.to_string() == own_Address.to_string() {
            succ_succ=own_Address
            return nil
        }
        conn, err = net.Dial("tcp", successor.Ip + ":" + successor.Port)
        if err != nil {
        fmt.Println("err in Stabilize 0", err)
        return err
        }
    }
    client := jsonrpc.NewClient(conn)
    var succ_pred Address
    err = client.Call("KeySpace.GetPredeccesor", "", &succ_pred)
    if err != nil {
        fmt.Println("err in Stabilize 1", err)
        return err
    }
    if(succ_pred.Ip!="" && in_between(own_Address.to_string(),succ_pred.to_string(),successor.to_string())) {
        successor = succ_pred
    }
    if prev_successor!=successor{
        fmt.Println("successor changed to - ",successor.to_string())
        conn.Close()
        conn, err = net.Dial("tcp", successor.Ip + ":" + successor.Port)
        if err != nil {
            fmt.Println("err in Stabilize 2", err)
            return err
        }
        client = jsonrpc.NewClient(conn)
    }
    var reply Address
    err = client.Call("KeySpace.GetSuccessor", own_Address.to_string(), &succ_succ)
    if err != nil {
        log.Fatal("Succ_succ not found error:", err)
    }
    if(prev_succ_succ!=succ_succ){
        fmt.Println("succ_succ changed to - ",succ_succ.to_string())
    }
    // fmt.Println("my succ_succ - ",succ_succ.to_string())

    // fmt.Println("calling notify to ",successor.to_string())
    err = client.Call("KeySpace.Notify", own_Address, &reply)
    if err != nil {
        conn.Close()
        fmt.Println("err in Stabilize 3", err)
        return err
    }
    conn.Close()
    return nil
}

func as_server_for_others() {
    var l net.Listener
    var e error
    if os.Args[1]=="create" {
        l, e = net.Listen("tcp", ":" + os.Args[2])
    } else {
        l, e = net.Listen("tcp", ":" + os.Args[3])
    }
    if e != nil {
        log.Fatal("listen error:", e)
    }
    keySpace2 := new(KeySpace)
    rpc.Register(keySpace2)
    for {
        conn, err := l.Accept()
        if err != nil {
          log.Printf("accept error: %s", conn)
          continue
        }
        // log.Printf("connection started: %v", conn.RemoteAddr())
        go jsonrpc.ServeConn(conn)
    }
}