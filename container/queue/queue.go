package queue

type Queue interface {
	Len() int
	Enqueue(interface{})
	Dequeue() interface{}
}
