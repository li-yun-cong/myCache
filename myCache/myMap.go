package myCache

import (
	"sync"
	"time"
)

// MyCache 自定义缓存结构体,使用泛型支持多类型
type MyCache[K comparable, V any] struct {
	items map[K]*Item[V]
	mu    sync.RWMutex

	GetDataTime func(cacheData V) time.Time // 从缓存数据中获取数据时间的匿名函数
}

// Item 自定义缓存类型
type Item[V any] struct {
	Value  V             // 缓存数据类型，支持任意类型
	Expire time.Duration // 过期时间TTL，单位s
}

// NewCache 创建缓存实例，使用指针，方便修改缓存对象和具体缓存内容的值，使用时需要注意影响
func NewCache[K comparable, V any](getDataTime func(cacheData V) time.Time) *MyCache[K, V] {
	if getDataTime == nil {
		// 返回一个默认函数，忽略传入的 cacheData
		getDataTime = func(cacheData V) time.Time {
			return time.Now()
		}
	}
	return &MyCache[K, V]{
		items:       make(map[K]*Item[V]), // 正确
		GetDataTime: getDataTime,
	}
}

// Set 设置缓存，expire可选，不传或传负数表示永不过期
func (c *MyCache[K, V]) Set(key K, value V, expire ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &Item[V]{
		Value: value,
	}

	if len(expire) > 0 && expire[0] >= 0 {
		c.items[key].Expire = expire[0]
	} else {
		c.items[key].Expire = -1
	}

}

// Get 获取缓存值，返回值和是否存在
func (c *MyCache[K, V]) Get(key K) (V, int) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	// 如果不存在，返回零值和false
	if !exists {
		var zero V
		// -1 代表数据找不到
		return zero, -1
	}

	// 检查是否过期
	if item.Expire >= 0 {
		c.mu.Lock()
		defer c.mu.Unlock()

		// 获取当前时间
		dataTime := c.GetDataTime(item.Value)
		nowTime := time.Now()
		if nowTime.Sub(dataTime) > item.Expire {
			var zero V
			// 0 代表数据过期
			return zero, 0
		}
	}
	// 1代表数据找到且在有效时间内
	return item.Value, 1
}
