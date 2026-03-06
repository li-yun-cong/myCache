package myCache

import (
	"errors"
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

// Update 更新缓存，不允许重新设置数据过期时间
func (c *MyCache[K, V]) Update(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &Item[V]{
		Value: value,
	}
}

// Get 获取缓存值，返回值和是否存在
func (c *MyCache[K, V]) Get(key K) (V, error) {
	c.mu.RLock()
	defer c.mu.Unlock()
	item, exists := c.items[key]

	// 如果不存在，返回零值和false
	if !exists {
		var zero V
		// -1 代表数据找不到
		return zero, errors.New("缓存中不存在指定键")
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
			return zero, errors.New("缓存中指定键数据已经过期")
		}
	}
	// 1代表数据找到且在有效时间内
	return item.Value, nil
}

// GetAll 获取全部缓存值
func (c *MyCache[K, V]) GetAll() map[K]V {
	c.mu.RLock()
	defer c.mu.Unlock()
	var allDatas map[K]V
	allDatas = make(map[K]V, len(c.items))
	for key, value := range c.items {
		allDatas[key] = value.Value
	}
	return allDatas
}

// Delete 删除指定键
func (c *MyCache[K, V]) Delete(key K) (err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.items[key]
	// 如果不存在，返回
	if !exists {
		return errors.New("缓存中不存在指定键")
	}
	delete(c.items, key)
	return nil
}

// DeleteByKeys 批量删除指定键
func (c *MyCache[K, V]) DeleteByKeys(keys []K) (deleteKeys []K, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, key := range keys {
		_, exists := c.items[key]
		// 如果不存在，返回
		if !exists {
			continue
		}
		delete(c.items, key)
		deleteKeys = append(deleteKeys, key)
	}
	return deleteKeys, nil
}

// Clear 清空缓存
func (c *MyCache[K, V]) Clear() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.items = make(map[K]*Item[V])
}
