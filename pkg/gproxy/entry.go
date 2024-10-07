package gproxy

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/slices"
)

func Entry(cfg *Config) error {

	mutex := new(sync.Mutex)

	policyFactories := CreatePolicyFactories()

	ctx, cancelFunc := context.WithCancel(context.Background())

	activeListeners := make(map[int]*Listener) // port: listener

	initiateListener := func(upstream *ListenerCfg) error {
		policyFactory, err := policyFactories(upstream.Policy)
		if err != nil {
			cancelFunc()
			return err
		}
		l := NewListener(ctx, upstream.Protocol, upstream.Port, upstream.Backends, policyFactory())
		mutex.Lock()
		activeListeners[upstream.Port] = l
		mutex.Unlock()
		go func() {
			err := l.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()
		return nil
	}

	updateConfig := func(newConfig *Config) error {
		// check listner diffs
		// delete listeners
		newListernersCfgMap := slices.ToMap(newConfig.Upstreams, func(l *ListenerCfg) (int, *ListenerCfg) {
			return l.Port, l
		})

		mutex.Lock()
		for port, listener := range activeListeners {
			if newListernersCfgMap[port] == nil {
				// stop and delete that listener
				listener.closeFunc()
				delete(activeListeners, port)
			}
		}
		mutex.Unlock()

		// create and add new listeners
		for port, listenerCfg := range newListernersCfgMap {
			if listener := activeListeners[port]; listener != nil {
				// update backends
				if err := listener.UpdateBackends(listenerCfg.Backends); err != nil {
					logging.GlobalLogger.Errorf(ctx, "failed to update backends for listener %d: %v", port, err)
					return err
				}
			} else {
				// up listener
				if err := initiateListener(listenerCfg); err != nil {
					return err
				}
			}
		}

		// update cfg
		mutex.Lock()
		cfg = newConfig
		mutex.Unlock()

		return nil
	}

	// initial load
	for _, upstream := range cfg.Upstreams {
		if err := initiateListener(upstream); err != nil {
			return err
		}
	}

	http.HandleFunc("/cfg", func(w http.ResponseWriter, r *http.Request) {
		// we can either POST or GET
		switch r.Method {
		case "PUT":
			data, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error parsing config: " + err.Error()))
				return
			}
			newCfg, err := UnmarshalConfig(data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error parsing config: " + err.Error()))
				return
			}
			if err = updateConfig(newCfg); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error updating config: " + err.Error()))
				return
			}
			w.Write(data)
		case "GET":
			data, err := MarshalConfig(cfg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error marshalling config: " + err.Error()))
				return
			}
			w.Write(data)
		}
	})

	go func() {
		if err := http.ListenAndServe("0.0.0.0:17768", nil); err != nil {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	return nil
}
