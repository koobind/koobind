/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// THIS HAS BEEN COPIED FROM
// csigs.k8s.io/controller-runtime@v0.5.0/pkg/webhook/internal/certwatcher/certwatcher.go
// AS IT WAS 'internal'
// Updated to:
// https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.12.3/pkg/certwatcher/certwatcher.go
// TODO: Use it as it is no not internal anymore

package certwatcher

import (
	"context"
	"crypto/tls"
	"github.com/go-logr/logr"
	"sync"

	"gopkg.in/fsnotify.v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// CertWatcher watches certificate and key files for changes.  When either file
// changes, it reads and parses both and calls an optional callback with the new
// certificate.
type CertWatcher struct {
	sync.Mutex

	logger   logr.Logger
	certPath string
	keyPath  string

	currentCert *tls.Certificate
	watcher     *fsnotify.Watcher
}

// New returns a new CertWatcher watching the given certificate and key.
func New(name, certPath, keyPath string) (*CertWatcher, error) {
	var err error

	cw := &CertWatcher{
		logger:   logf.Log.WithName("certwatcher").WithValues("name", name, "certPath", certPath, "keyPath", keyPath),
		certPath: certPath,
		keyPath:  keyPath,
	}

	// Initial read of certificate and key.
	if err := cw.ReadCertificate(); err != nil {
		return nil, err
	}

	cw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return cw, nil
}

// GetCertificate fetches the currently loaded certificate, which may be nil.
func (cw *CertWatcher) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cw.Lock()
	defer cw.Unlock()
	return cw.currentCert, nil
}

// Start starts the watch on the certificate and key files.
func (cw *CertWatcher) Start(ctx context.Context) error {
	files := []string{cw.certPath, cw.keyPath}

	for _, f := range files {
		if err := cw.watcher.Add(f); err != nil {
			return err
		}
	}

	go cw.Watch()

	cw.logger.Info("Starting cert. watcher")

	// Block until the stop channel is closed.
	<-ctx.Done()

	return cw.watcher.Close()
}

// Watch reads events from the watcher's channel and reacts to changes.
func (cw *CertWatcher) Watch() {
	for {
		select {
		case event, ok := <-cw.watcher.Events:
			// Channel is closed.
			if !ok {
				return
			}

			cw.handleEvent(event)

		case err, ok := <-cw.watcher.Errors:
			// Channel is closed.
			if !ok {
				return
			}

			cw.logger.Error(err, "certificate watch error")
		}
	}
}

// ReadCertificate reads the certificate and key files from disk, parses them,
// and updates the current certificate on the watcher.  If a callback is set, it
// is invoked with the new certificate.
func (cw *CertWatcher) ReadCertificate() error {
	cert, err := tls.LoadX509KeyPair(cw.certPath, cw.keyPath)
	if err != nil {
		return err
	}

	cw.Lock()
	cw.currentCert = &cert
	cw.Unlock()

	cw.logger.Info("Updated current TLS certificate")

	return nil
}

func (cw *CertWatcher) handleEvent(event fsnotify.Event) {
	// Only care about events which may modify the contents of the file.
	if !(isWrite(event) || isRemove(event) || isCreate(event)) {
		return
	}

	cw.logger.V(1).Info("certificate event", "event", event)

	// If the file was removed, re-add the watch.
	if isRemove(event) {
		if err := cw.watcher.Add(event.Name); err != nil {
			cw.logger.Error(err, "error re-watching file")
		}
	}

	if err := cw.ReadCertificate(); err != nil {
		cw.logger.Error(err, "error re-reading certificate")
	}
}

func isWrite(event fsnotify.Event) bool {
	return event.Op&fsnotify.Write == fsnotify.Write
}

func isCreate(event fsnotify.Event) bool {
	return event.Op&fsnotify.Create == fsnotify.Create
}

func isRemove(event fsnotify.Event) bool {
	return event.Op&fsnotify.Remove == fsnotify.Remove
}
