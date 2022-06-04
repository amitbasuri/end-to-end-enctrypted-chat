# End-to-end encrypted chat

### Solution Approach
client generates key pair and shares public key with server, such that any other client
can access public key.

if Alice want to send message to Bob.
Alice fetches public key of Bob from server.
Alice Client encrypts the message alice wants to send to bob using Bob's public key
Bob decrypts that message using its private key.

### Example Usage
### Run Server
 ```shell
 go build  -o ./chatServer ./server
 ./chatServer
 ```
### Run Client
 ```shell
 go build  -o ./clientBuild ./client
 ./clientBuild
  ```

### Sample Client - Alice
 ```text
 go build  -o ./clientBuild ./client
 ./clientBuild
Enter your name: alice
Enter name of user to send Message to: bob
Enter Message: hey bob
{"message":"message sent"}
Enter name of user to send Message to: 
 Received messages --> 
From: bob, Message: hii Alice :)
 ```
### Sample Client - Bob
 ```text
 go build  -o ./clientBuild ./client
 ./clientBuild
Enter your name: bob
Enter name of user to send Message to: 
 Received messages --> 
From: alice, Message: hey bob
From: eric, Message: Hi, this is Eric
alice
Enter Message: hii Alice :)
{"message":"message sent"}
Enter name of user to send Message to: 
 Received messages --> 
From: alice, Message: how r u?
 ```
### Run Tests
 ```shell
 go test ./...
 ```