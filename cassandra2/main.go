package main

import (
    "bufio"
    "log"
    "strings"
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
    "time"
    )


var (
    successor Address = Address{"",""}
    predecessor Address = Address{"",""}
    own_Address = Address{"",""}
    store map[string]string = make(map[string]string)
    store_pred map[string]string = make(map[string]string)
    onlyOne bool
    keySpace1 = new(KeySpace)
    succ_succ Address = Address{"",""}
)

func (t *KeySpace)CallInsert(keyVal KeyVal, reply *string) error{
    if successor==own_Address {
        fmt.Println("successor == ownaddress = ",successor.to_string())
        var reply_str string
        err := keySpace1.Insert(keyVal, &reply_str)
        *reply = reply_str
        return err
    } else {
        var to_send Address 
        err := keySpace1.FindSuccessor(keyVal.Key, &to_send)
        if err != nil {
            fmt.Println("error in FindSuccessor 1", err)
            return err
        }
        if to_send==own_Address {
            fmt.Println("to_send == ownaddress = ",own_Address.to_string())
            var reply_str string
            err := keySpace1.Insert(keyVal, &reply_str)
            *reply = reply_str
            return err
        } else {
            fmt.Println("to_send = ",to_send.to_string())
            conn, err := net.Dial("tcp", to_send.Ip + ":" + to_send.Port)
            if err != nil {
                return err
            }
            client := jsonrpc.NewClient(conn)
            var reply_str string
            err = client.Call("KeySpace.Insert", keyVal, &reply_str)
            *reply = reply_str
            if err != nil {
                return err
            }
            conn.Close()
        }
    }
    return nil
}

func (t *KeySpace)CallRemove(key string, reply *string) error{
    if successor==own_Address {
        fmt.Println("successor == ownaddress = ",successor.to_string())
        err := keySpace1.Remove(key,reply)
        return err
    } else {
        var to_send Address 
        err := keySpace1.FindSuccessor(key, &to_send)
        if err != nil {
            return err
        }
        if to_send==own_Address {
            fmt.Println("to_send == ownaddress = ",to_send.to_string())
            err := keySpace1.Remove(key,reply)
            return err
        } else {
            fmt.Println("to_send = ",to_send.to_string())
            conn, err := net.Dial("tcp", to_send.Ip + ":" + to_send.Port)
            if err != nil {
                return err
            }
            client := jsonrpc.NewClient(conn)
            err = client.Call("KeySpace.Remove", key, reply)
            if err != nil {
                return err
            }
            conn.Close()
        }
    }
    return nil
}


func (t *KeySpace)CallGet(key string, val *string) error{
    if successor==own_Address {
        fmt.Println("successor == ownaddress = ",successor.to_string())
        var val_str string
        err := keySpace1.Get(key,&val_str)
        *val = val_str
        return err
    } else {
        var to_send Address 
        err := keySpace1.FindSuccessor(key, &to_send)
        if err != nil {
            return err
        }
        if to_send==own_Address {
            fmt.Println("to_send == ownaddress = ",to_send.to_string())
            var val_str string
            err := keySpace1.Get(key,&val_str)
            *val = val_str
            return err
        } else {
            fmt.Println("to_send = ",to_send.to_string())
            conn, err := net.Dial("tcp", to_send.Ip + ":" + to_send.Port)
            if err != nil {
                return err
            }
            client := jsonrpc.NewClient(conn)
            var val_str string
            err = client.Call("KeySpace.Get", key, &val_str)
            *val = val_str
            if err != nil {
                return err
            }
            conn.Close()
        }
    }
    return nil
}

func callStabilize() {
    for _ = range time.NewTicker(1 * time.Second).C {
        Stabilize()
    }
}

func main() {
    //./main create [portToListen]
    // ./main [ip_someNode] [port_someNode] [portToListen]
    rpc.Register(keySpace1)
    go as_server_for_others()

    if os.Args[1]!="create" {
        conn, err := net.Dial("tcp", os.Args[1]+":"+os.Args[2])
        if err != nil {
            log.Fatal("Connectiong:", err)
        }
        client := jsonrpc.NewClient(conn)
        ip,err := externalIP()
        own_Address = Address{ip,os.Args[3]}
        fmt.Println("address:- ",own_Address.to_string()," , Hash of address - ", hash(own_Address.to_string()))
        err = client.Call("KeySpace.FindSuccessor", own_Address.to_string(), &successor)
        if err != nil {
            log.Fatal("Successor not found error:", err)
        }
        fmt.Println("my successor - ",successor.to_string())
        conn.Close()
        conn, err = net.Dial("tcp", successor.Ip+":"+successor.Port)
         if err != nil {
            log.Fatal("Connectiong:", err)
        }
        client = jsonrpc.NewClient(conn)
        err = client.Call("KeySpace.Notify", own_Address, nil)
        if err != nil {
            log.Fatal("error while notifying:", err)
        }
        err = client.Call("KeySpace.Getkeyval", own_Address, &store)
        if err != nil {
            log.Fatal("error while getting keyval:", err)
        }
        err = client.Call("KeySpace.GetSuccessor", own_Address.to_string(), &succ_succ)
        if err != nil {
            log.Fatal("Succ_succ not found error:", err)
        }
        fmt.Println("my succ_succ - ",succ_succ.to_string())
        conn.Close()

    } else { 
        ip,_ := externalIP()
        successor = Address{ip,os.Args[2]}
        succ_succ = Address{ip,os.Args[2]}
        own_Address = Address{ip,os.Args[2]}
        fmt.Println("address:- ",own_Address.to_string()," , Hash of address - ", hash(own_Address.to_string()))
    }
    go callStabilize()

    for true {
        string_return := ""
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Enter Command(Insert/Remove/Get):") 
        text, _ := reader.ReadString('\n')
        text = strings.TrimSpace(text)
        // fmt.Println("hihi",text)
        if text=="Insert" { 
            var keyVal_obj KeyVal
            fmt.Print("Enter Key:") 
            text, _ := reader.ReadString('\n')
            text = strings.TrimSpace(text)
            keyVal_obj.Key = text
            fmt.Println("Hash of key - ",hash(text))
            fmt.Print("Enter Val:") 
            text, _ = reader.ReadString('\n')
            text = strings.TrimSpace(text)
            keyVal_obj.Val = text
            fmt.Println(keyVal_obj)
            err := keySpace1.CallInsert(keyVal_obj, &string_return)
            if err != nil {
                fmt.Println("error:", err)
            } else {
                fmt.Println("KeyVal Inserted")
            }
        } else if text=="Remove" {
            fmt.Print("Enter key:")
            text, _ := reader.ReadString('\n')
            text = strings.TrimSpace(text)

            err := keySpace1.CallRemove(text, &string_return)
            if err != nil {
                fmt.Println("error:", err)
            } else {
                fmt.Println("Key Removed")
            }
            
        } else if text=="Get" {
            fmt.Print("Enter key_string:")
            text, _ := reader.ReadString('\n')
            text = strings.TrimSpace(text)
            err := keySpace1.CallGet(text, &string_return)
            if err != nil {
                fmt.Println("error:", err)
            } else if string_return==""{
                fmt.Println("error:No val")
            } else {
                fmt.Println(string_return)
            }
        } else if text=="Print" {
            fmt.Println("My own_Address = ",own_Address.to_string(),"Hash = ",hash(own_Address.to_string()))
            fmt.Println("My successor = ",successor.to_string(),"Hash = ",hash(successor.to_string()))
            fmt.Println("My predecessor = ",predecessor.to_string(),"Hash = ",hash(predecessor.to_string()))
            fmt.Println("My succ_succ = ",succ_succ.to_string(),"Hash = ",hash(succ_succ.to_string()))
        } else if text=="Print_dict" {
            fmt.Println("store")
            for key, value := range store {
                fmt.Println("Key:", key, "Value:", value)
            }
        } else if text=="Print_dictpred" {
            fmt.Println("store_pred")
            for key, value := range store_pred {
                fmt.Println("Key:", key, "Value:", value)
            }
        }
    }
}