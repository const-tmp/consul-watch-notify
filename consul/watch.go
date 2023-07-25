package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type Watcher[K comparable, V any] struct {
	Address, Token string
	KeyGetter      func(V) K
	ChangeDetector func(prev V, cur V) bool
	Notifiers      []NotifyConfig
}

func (w Watcher[K, V]) getMapFromSlice(slice []V) map[K]V {
	newMap := make(map[K]V)
	for _, v := range slice {
		newMap[w.KeyGetter(v)] = v
	}
	return newMap
}

func (w Watcher[K, V]) HandlerFactory() watch.HandlerFunc {
	state := make(map[K]V)

	return func(u uint64, i any) {
		var newState map[K]V

		switch v := i.(type) {
		case []V:
			newState = w.getMapFromSlice(v)
		case map[K]V:
			newState = v
		default:
			panic(fmt.Sprintf("unknown type: %T", i))
		}

		for k, v := range state {
			if _, ok := newState[k]; !ok {
				delete(state, k)
				w.notifyDeregister(k, v)
			}
		}

		for k, v := range newState {
			if prev, ok := state[k]; ok {
				state[k] = v
				if w.ChangeDetector != nil && w.ChangeDetector(prev, v) {
					w.notifyChange(k, v)
				}
			} else {
				state[k] = v
				w.notifyRegister(k, v)
			}
		}
	}
}

// CompareSlice returns true if two slices are equal
func CompareSlice[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SlicesNotEqual returns true if two slices are not equal
func SlicesNotEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return true
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return true
		}
	}
	return false
}

func NewServicesWatcher() Watcher[string, []string] {
	return Watcher[string, []string]{
		ChangeDetector: func(prev []string, cur []string) bool { return SlicesNotEqual[string](prev, cur) },
	}
}

func NewNodesWatcher() Watcher[string, *api.Node] {
	return Watcher[string, *api.Node]{
		KeyGetter: func(c *api.Node) string { return c.Node },
		ChangeDetector: func(prev *api.Node, cur *api.Node) bool {
			return cur.Node != prev.Node || cur.Address != prev.Address || cur.ID != prev.ID
		},
	}
}

func NewChecksWatcher() Watcher[string, *api.HealthCheck] {
	return Watcher[string, *api.HealthCheck]{
		KeyGetter:      func(c *api.HealthCheck) string { return c.CheckID },
		ChangeDetector: func(prev *api.HealthCheck, cur *api.HealthCheck) bool { return cur.Status != prev.Status },
	}
}
