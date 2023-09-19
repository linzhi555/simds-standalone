package main

import "errors"

type MessageBody interface {
	MessageBody()
}

func (TaskInfo) MessageBody() {}
func (NodeInfo) MessageBody() {}

type Message struct {
	From     string
	To       string
	Content  string
	LeftTime int32
	Body     MessageBody
}

type MessageQueue struct {
	buffers []Message
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		buffers: make([]Message, 0),
	}
}

func (mch *MessageQueue) Empty() bool {
	return mch.Len() == 0

}
func (mch *MessageQueue) Len() int {
	return len(mch.buffers)
}

func (mch *MessageQueue) InQueue(m Message) {
	mch.buffers = append(mch.buffers, m)
}

func (mch *MessageQueue) Dequeue() (Message, error) {
	if mch.Empty() == true {
		return Message{}, errors.New("the queue is Empty")
	}
	res := mch.buffers[0]
	mch.buffers = mch.buffers[1:mch.Len()]
	return res, nil
}
