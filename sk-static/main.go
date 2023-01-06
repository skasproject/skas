

package main

import "fmt"
import "skas/sk-common/proto"


func main() {
    fmt.Println("Allo...")

    req := &proto.LoginRequest{
        Client: "xxxxx",

    }
    fmt.Printf("%s\n", req.Client)
}

