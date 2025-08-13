// config/watcher.go
package config

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// 配置管理器（用于动态更新配置）
type Manager struct {
	cfg      *Config
	path     string       // 配置文件路径
	mu       sync.RWMutex // 保证配置读写安全
	watcher  *fsnotify.Watcher
	stopChan chan struct{} // 停止信号
}

// 创建配置管理器
func NewManager(path string) (*Manager, error) {
	// 初始加载配置
	cfg, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	// 初始化文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监听器失败: %w", err)
	}

	// 监听配置文件所在目录（避免文件被替换后监听失效）
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("监听配置文件失败: %w", err)
	}

	m := &Manager{
		cfg:      cfg,
		path:     path,
		watcher:  watcher,
		stopChan: make(chan struct{}),
	}

	// 启动监听协程
	go m.watch()

	return m, nil
}

// 监听配置文件变化并热重载
func (m *Manager) watch() {
	ticker := time.NewTicker(m.cfg.ReloadInterval) // 定期检查（防止某些场景下事件丢失）
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			// 仅处理配置文件的写事件或删除后重建事件
			if event.Name == m.path && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				m.reload()
			}
		case <-ticker.C:
			m.reload() // 定期检查更新
		case <-m.stopChan:
			m.watcher.Close()
			return
		}
	}
}

// 重新加载配置
func (m *Manager) reload() {
	newCfg, err := LoadFromFile(m.path)
	if err != nil {
		fmt.Printf("动态重载配置失败: %v\n", err)
		return
	}

	// 安全替换配置（写锁保证原子性）
	m.mu.Lock()
	m.cfg = newCfg
	m.mu.Unlock()

	fmt.Println("配置已动态更新")
}

// 获取当前配置（读锁保证并发安全）
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg // 注意：返回的是指针，避免外部修改，可考虑返回深拷贝
}

// 停止监听
func (m *Manager) Stop() {
	close(m.stopChan)
}
