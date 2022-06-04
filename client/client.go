package main

import (
	"bufio"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

const (
	server = "http://127.0.0.1:8888"
)

func main() {
	svc := NewService()
	go svc.ReceiveMessage()
	svc.SendMessage()
}

type service struct {
	// userName is name of the user
	userName string

	// client is http client which connects to server
	client *resty.Client

	// apiKey is api key for sending and receiving messages
	apiKey string

	// privateKey is private key of user
	privateKey *rsa.PrivateKey

	// pubKey is public key of user
	pubKey *rsa.PublicKey

	// usersPubKeyCache stores all users public key
	usersPubKeyCache map[string]*rsa.PublicKey

	// usersPubKeyCacheLock is used read/write to usersPubKeyCache
	usersPubKeyCacheLock sync.RWMutex
}

func NewService() *service {

	s := &service{
		client:               resty.New(),
		usersPubKeyCache:     map[string]*rsa.PublicKey{},
		usersPubKeyCacheLock: sync.RWMutex{},
	}
	err := s.createUser()
	if err != nil {
		log.Fatal(err.Error())
	}
	return s
}

// getUserPubKey returns public key of user
// if pub key is not found in cache it is fetched from server
func (svc *service) getUserPubKey(user string) *rsa.PublicKey {
	svc.usersPubKeyCacheLock.RLock()
	key, ok := svc.usersPubKeyCache[user]
	svc.usersPubKeyCacheLock.RUnlock()
	if ok {
		return key
	}

	resp, _ := svc.client.R().
		SetResult(&GetPubKeySuccessResponse{}).
		SetError(&GetPubKeyErrorResponse{}).
		Get(fmt.Sprintf("%s/user/%s", server, user))
	if resp.IsError() {
		log.Fatal(resp.Error().(*GetPubKeyErrorResponse).Message)
	}
	pubKeyBytes := resp.Result().(*GetPubKeySuccessResponse).PubKey
	pubKey := BytesToPublicKey(pubKeyBytes)

	svc.usersPubKeyCacheLock.Lock()
	svc.usersPubKeyCache[user] = pubKey
	svc.usersPubKeyCacheLock.Unlock()

	return pubKey

}

// createUser creates a user in server
func (svc *service) createUser() error {
	u := &CreateUserResponse{}
	fmt.Print("Enter your name: ")
	var name string
	fmt.Scanln(&name)
	svc.userName = name
	svc.privateKey, svc.pubKey = GenerateKeyPair()
	svc.usersPubKeyCache[name] = svc.pubKey

	resp, _ := svc.client.R().
		SetBody(map[string]interface{}{"name": name, "pubKey": PublicKeyToBytes(svc.pubKey)}).
		SetResult(u).
		SetError(u).
		Post(fmt.Sprintf("%s/user", server))
	if resp.IsError() {
		return errors.New(resp.Error().(*CreateUserResponse).Message)
	}

	svc.apiKey = resp.Result().(*CreateUserResponse).ApiKey
	return nil
}

// SendMessage sends message to user
// messages are encrypted with recipient's public key
func (svc *service) SendMessage() {

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter name of user to send Message to: ")
		scanner.Scan()
		to := scanner.Text()

		fmt.Print("Enter Message: ")
		scanner.Scan()
		message := scanner.Text()

		toUserPubKey := svc.getUserPubKey(to)

		encryptedMessage := EncryptWithPublicKey([]byte(message), toUserPubKey)
		resp, _ := svc.client.R().
			SetHeader("X-API-Key", svc.apiKey).
			SetBody(map[string]interface{}{"to": to, "msg": encryptedMessage}).
			Post(fmt.Sprintf("%s/message", server))
		if resp.IsError() {
			fmt.Println(resp)
		}
	}
}

// ReceiveMessage checks for any messages on regular interval
// all received messages are decrypted using private key
func (svc *service) ReceiveMessage() {
	for {
		<-time.After(time.Second * 5)
		resp, _ := svc.client.R().
			SetHeader("X-API-Key", svc.apiKey).
			SetResult(&UserMessages{}).
			Get(fmt.Sprintf("%s/message", server))
		messages := resp.Result().(*UserMessages).Messages
		if len(messages) > 0 {
			fmt.Println("\n Received messages --> ")
			for _, message := range messages {
				decryptMsg := string(DecryptWithPrivateKey(message.Message, svc.privateKey))
				fmt.Printf("From: %s, Message: %s\n", message.From, decryptMsg)
			}
		}
	}
}

type CreateUserResponse struct {
	Message string `json:"message"`
	ApiKey  string `json:"api_key"`
}

type GetPubKeySuccessResponse struct {
	PubKey []byte `json:"pubKey"`
}

type GetPubKeyErrorResponse struct {
	Message string `json:"message"`
}

type UserMessages struct {
	Messages []UserMessage
}

type UserMessage struct {
	From    string
	Message []byte
}
