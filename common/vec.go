package common

import "errors"

// Vec 为三种类型定义Vector
type Vec[T any] []T

//// MessageBody Vec[T] 是 MessageBody
//func (vec Vec[T]) MessageBody() {}

// InQueueFront 在Vector头部入队
func (vec *Vec[T]) InQueueFront(data T) {
	*vec = append(Vec[T]{data}, *vec...)

}

// Clone 拷贝一份新的Vec[T]
func (vec *Vec[T]) Clone() *Vec[T] {
	newVec := make(Vec[T], len(*vec))
	copy(newVec, *vec)
	return &newVec
}

// InQueue 在Vector尾部入队
func (vec *Vec[T]) InQueueBack(data T) {
	*vec = append(*vec, data)
}

// Len 返回Vector 长度
func (vec *Vec[T]) Len() int {
	return len(*vec)
}

// Empty 返回Vector 是否为空
func (vec *Vec[T]) Empty() bool {
	return vec.Len() == 0
}

// Dequeue 在Vector头部出队
func (vec *Vec[T]) Dequeue() (T, error) {
	var res T
	if vec.Empty() {
		return res, errors.New("the queue is Empty")
	}
	res = (*vec)[0]
	*vec = (*vec)[1:vec.Len()]
	return res, nil
}

// Dequeue 在Vector尾部出队
func (vec *Vec[T]) Pop() (T, error) {
	var res T
	if vec.Empty() {
		return res, errors.New("the queue is Empty")
	}
	res = (*vec)[vec.Len()-1]
	*vec = (*vec)[0 : vec.Len()-1]
	return res, nil
}

// Delete 删除Vector的某个元素
func (vec *Vec[T]) Delete(index int) {
	if index >= vec.Len() || index < 0 {
		panic("index out of range")
	} else if index == vec.Len()-1 {
		*vec = (*vec)[:vec.Len()-1]
	} else {
		copy((*vec)[index:], (*vec)[index+1:])
		*vec = (*vec)[:vec.Len()-1]
	}
}

// 清空Vector
func (vec *Vec[T]) Clean() {
	*vec = (*vec)[0:0]
}
