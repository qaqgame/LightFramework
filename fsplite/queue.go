package fsplite

type (
	//Queue 队列
	Queue struct {
		top    *node
		rear   *node
		length int
	}
	//双向链表节点
	node struct {
		pre   *node
		next  *node
		value interface{}
	}
)

// NewQueue : Create a new queue
func NewQueue() *Queue {
	return &Queue{nil, nil, 0}
}

// Len : 获取队列长度
func (queue *Queue) Len() int {
	return queue.length
}

// Any : 返回true队列不为空
func (queue *Queue) Any() bool {
	return queue.length > 0
}

// Peek : 返回队列顶端元素
func (queue *Queue) Peek() interface{} {
	if queue.top == nil {
		return nil
	}
	return queue.top.value
}

// Push : 入队操作
func (queue *Queue) Push(v interface{}) {
	n := &node{nil, nil, v}
	if queue.length == 0 {
		queue.top = n
		queue.rear = queue.top
	} else {
		n.pre = queue.rear
		queue.rear.next = n
		queue.rear = n
	}
	queue.length++
}

// Pop : 出队操作
func (queue *Queue) Pop() interface{} {
	if queue.length == 0 {
		return nil
	}
	n := queue.top
	if queue.top.next == nil {
		queue.top = nil
	} else {
		queue.top = queue.top.next
		queue.top.pre.next = nil
		queue.top.pre = nil
	}
	queue.length--
	return n.value
}

// Contain : 是否含有
func (queue *Queue) Contain(v interface{}) bool {
	if queue.Len() == 0 {
		return false
	}

	n := queue.top
	for n != nil {
		if n.value == v {
			return true
		}
		n = n.next
	}
	return false
}

// Clear : clear content
func (queue *Queue) Clear() {
	queue.top = nil
	queue.rear = nil
	queue.length = 0
}
