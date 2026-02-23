// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package authkit

import (
	"fmt"
	"sync"

	"go.xiexianbin.cn/authkit/types"
)

var (
	mu            sync.RWMutex
	registry      = make(map[string]types.Provider)
	providerNames = make([]string, 0)
)

// RegisterProvider registers a new OAuth provider.
// It is thread-safe and can be called at init or runtime.
func RegisterProvider(name string, p types.Provider) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[name]; !exists {
		providerNames = append(providerNames, name)
	}
	registry[name] = p
}

// GetProvider get an OAuth Provider instance by name
func GetProvider(name string) (types.Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	provider, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", name)
	}
	return provider, nil
}

// GetProviders returns a list of registered provider names in order of registration
func GetProviders() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, len(providerNames))
	copy(names, providerNames)
	return names
}
