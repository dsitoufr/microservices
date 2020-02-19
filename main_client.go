package main

import (
"fmt"
"flag"
"time"
"crypto/sha256"
"encoding/hex"
"myexamples/chat/proto"
"log"
"google.golang.org/grpc"
"golang.org/x/net/context"
"sync"
"bufio"
"os"
)


// declarations

var client proto.BroadcastClient

var wait *sync.WaitGroup

// functions 

func init() {
   wait = &sync.WaitGroup{} 
}


//main

func main() {

  timestamp := time.Now()
  done := make(chan int)

  // read user name  
  name := flag.String("N","Anon","The name of the user")
  flag.Parse()
  id := sha256.Sum256([]byte(timestamp.String() + *name))

  //create user object
  user := &proto.User{
            Id: hex.EncodeToString(id[:]),
            Name: *name,
          }

  //connect to server
  conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
  if err != nil {
     log.Fatalf("Could not connect to server: %v", err)
  }
  
  //create client api 
  client = proto.NewBroadcastClient(conn)

  //create stream client to receive msg 
  stream, err := client.CreateStream(context.Background(), &proto.Connect{
              User: user,
              Active: true, 
  })

  //reception continue des msg
  wait.Add(1)
  go func() {
      defer wait.Done()
     
      for {
           msg, err := stream.Recv()
           if err != nil {
               fmt.Errorf("Error reading message stream: %v\n", err)
               break
           }
     
           fmt.Printf("%v : %s\n", msg.Id, msg.Content)
      }  
      
      
  }()

 //ecriture continue des msg
 wait.Add(1)
 go func() {
      defer wait.Done()
     
      scanner := bufio.NewScanner(os.Stdin)
      for scanner.Scan() {

          msg := &proto.Message{
                    Id: user.Id,
                    Content: scanner.Text(),
                    Timestamp: timestamp.String(), 
                 }

          _, err := client.BroadcastMessage(context.Background(), msg)
          if err != nil {
             fmt.Printf("Error sending message: %v\n", err)
             break
          }
      }
 }()

//synchro routines
go func() {
   wait.Wait()  //until all decrement
   close(done)
}()

 <-done

}
