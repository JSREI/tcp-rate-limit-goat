// Package utils 提供实用的数据结构和工具函数
// 包含循环缓冲区（环形缓冲区）的实现
package utils

import (
	"fmt"
)

// CircularBuffer 是一个固定大小的循环缓冲区（环形缓冲区）实现
// 特点：
// 1. 固定容量
// 2. 当缓冲区已满时，新元素会覆盖最旧的元素
// 3. 支持高效的入队和出队操作
type CircularBuffer struct {
	// buffer 存储实际数据的底层切片
	buffer []interface{}

	// head 指向缓冲区中最旧元素的索引
	head int

	// tail 指向下一个可插入元素的索引
	tail int

	// count 记录缓冲区中当前元素的数量
	count int

	// capacity 缓冲区的最大容量
	capacity int
}

// NewCircularBuffer 创建一个新的循环缓冲区
// 参数:
//   - capacity: 缓冲区的最大容量
//
// 返回值:
//   - *CircularBuffer: 初始化的循环缓冲区指针
func NewCircularBuffer(capacity int) *CircularBuffer {
	return &CircularBuffer{
		buffer:   make([]interface{}, capacity), // 创建固定大小的底层切片
		head:     0,                             // 初始头部索引为0
		tail:     0,                             // 初始尾部索引为0
		count:    0,                             // 初始元素数量为0
		capacity: capacity,                      // 设置缓冲区容量
	}
}

// Enqueue 向循环缓冲区添加新元素
// 如果缓冲区已满，最旧的元素将被新元素覆盖
// 参数:
//   - value: 要添加的元素（支持任意类型）
func (cb *CircularBuffer) Enqueue(value interface{}) {
	// 在尾部索引处存储新元素
	cb.buffer[cb.tail] = value

	// 更新尾部索引（循环）
	cb.tail = (cb.tail + 1) % cb.capacity

	// 如果缓冲区已满，移动头部索引
	if cb.count == cb.capacity {
		cb.head = (cb.head + 1) % cb.capacity
	} else {
		// 否则增加元素计数
		cb.count++
	}
}

// Dequeue 从循环缓冲区移除并返回最旧的元素
// 返回值:
//   - interface{}: 最旧的元素，如果缓冲区为空则返回nil
func (cb *CircularBuffer) Dequeue() interface{} {
	// 如果缓冲区为空，返回nil
	if cb.count == 0 {
		return nil
	}

	// 获取头部元素
	value := cb.buffer[cb.head]

	// 更新头部索引（循环）
	cb.head = (cb.head + 1) % cb.capacity

	// 减少元素计数
	cb.count--

	return value
}

// Elements 返回缓冲区中的所有元素（按顺序）
// 返回值:
//   - []interface{}: 包含所有元素的切片
func (cb *CircularBuffer) Elements() []interface{} {
	// 创建与当前元素数量相同大小的切片
	elements := make([]interface{}, cb.count)

	// 遍历并复制元素
	for i, j := 0, cb.head; i < cb.count; i, j = i+1, j+1 {
		j = j % cb.capacity
		elements[i] = cb.buffer[j]
	}

	return elements
}

// main 函数提供了循环缓冲区使用的示例
func main() {
	// 创建容量为5的循环缓冲区
	cb := NewCircularBuffer(5)

	// 添加元素
	cb.Enqueue(1)
	cb.Enqueue(2)
	cb.Enqueue(3)
	fmt.Println(cb.Elements()) // 输出: [1 2 3]

	// 移除元素并添加新元素
	fmt.Println(cb.Dequeue()) // 输出: 1
	cb.Enqueue(4)
	fmt.Println(cb.Elements()) // 输出: [2 3 4]

	// 继续移除元素
	fmt.Println(cb.Dequeue()) // 输出: 2
	fmt.Println(cb.Dequeue()) // 输出: 3
	fmt.Println(cb.Dequeue()) // 输出: 4
	fmt.Println(cb.Dequeue()) // 输出: <nil> 因为缓冲区已空
}
