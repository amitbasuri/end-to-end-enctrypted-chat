package main

import "sync"

type ReceiverMessage struct {
	From    string
	Message []byte
}

type SenderMessage struct {
	To      string
	Message []byte
}

type store interface {
	SetUserPubKey(user string, pubKey []byte)
	GetUserPubKey(user string) ([]byte, bool)

	SetApiKeyUser(apiKey string, user string)
	Authenticate(apiKey string) (string, bool)

	GetUserMessages(user string) []ReceiverMessage
	AddMessage(from string, msg SenderMessage)
}

func NewInMem() store {
	return &inMem{
		usersPubKeyMap: map[string][]byte{},
		pubKeyLock:     sync.RWMutex{},
		apiKeyUserMap:  map[string]string{},
		apiKeyLock:     sync.RWMutex{},
		messages:       map[string][]ReceiverMessage{},
		messagesLock:   sync.Mutex{},
	}
}

type inMem struct {
	// usersPubKeyMap stores public key of all users
	usersPubKeyMap map[string][]byte
	pubKeyLock     sync.RWMutex

	// apiKeyUserMap stores api-key --> user mapping, used for authentication
	apiKeyUserMap map[string]string
	apiKeyLock    sync.RWMutex

	// messages stores all the messages which are yet to be sent to recipient
	messages     map[string][]ReceiverMessage
	messagesLock sync.Mutex
}

func (i *inMem) SetUserPubKey(user string, pubKey []byte) {
	i.pubKeyLock.Lock()
	defer i.pubKeyLock.Unlock()

	i.usersPubKeyMap[user] = pubKey
}

func (i *inMem) GetUserPubKey(user string) ([]byte, bool) {
	i.pubKeyLock.RLock()
	defer i.pubKeyLock.RUnlock()

	key, ok := i.usersPubKeyMap[user]
	return key, ok
}

func (i *inMem) SetApiKeyUser(apiKey string, user string) {
	i.apiKeyLock.Lock()
	defer i.apiKeyLock.Unlock()

	i.apiKeyUserMap[apiKey] = user
}

func (i *inMem) Authenticate(apiKey string) (string, bool) {
	i.apiKeyLock.RLock()
	defer i.apiKeyLock.RUnlock()

	user, ok := i.apiKeyUserMap[apiKey]
	return user, ok
}

func (i *inMem) GetUserMessages(user string) []ReceiverMessage {
	i.messagesLock.Lock()
	defer i.messagesLock.Unlock()

	msgs, _ := i.messages[user]
	if msgs == nil || len(msgs) == 0 {
		return []ReceiverMessage{}
	}
	res := make([]ReceiverMessage, len(msgs))
	copy(res, msgs)
	i.messages[user] = []ReceiverMessage{}
	return res
}

func (i *inMem) AddMessage(from string, msg SenderMessage) {
	i.messagesLock.Lock()
	defer i.messagesLock.Unlock()

	i.messages[msg.To] = append(i.messages[msg.To], ReceiverMessage{
		From:    from,
		Message: msg.Message,
	})
}
