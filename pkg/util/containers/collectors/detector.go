// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package collectors

import (
	"errors"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/DataDog/datadog-agent/pkg/util/retry"
)

var ErrNothing = errors.New("No collector available")

type Detector struct {
	candidates        map[string]Collector
	detected          map[string]Collector
	preferedCollector Collector
	preferedName      string
}

func NewDetector() *Detector {
	d := &Detector{
		candidates: make(map[string]Collector),
		detected:   make(map[string]Collector),
	}
	// Load candidates from catalog
	for n, f := range DefaultCatalog {
		d.candidates[n] = f()
	}
	return d
}

func (d *Detector) GetPrefered() (Collector, string, error) {
	if d.candidates != nil {
		d.detectCandidates()
	}
	if d.preferedCollector != nil {
		return d.preferedCollector, d.preferedName, nil
	}
	return nil, "", ErrNothing
}

func (d *Detector) detectCandidates() {
	// Stop detection when all candidates are tested
	if len(d.candidates) == 0 {
		d.candidates = nil
		d.detected = nil
		return
	}

	foundNew := false
	// Retry all remaining candidates
	for name, c := range d.candidates {
		err := c.Detect()
		if retry.IsErrWillRetry(err) {
			log.Debugf("will retry collector %s later: %s", name, err)
			continue // we want to retry later
		}
		if err != nil {
			log.Debugf("%s collector cannot detect: %s", name, err)
			delete(d.candidates, name)
		} else {
			log.Infof("%s collector successfully detected", name)
			d.detected[name] = c
			foundNew = true
			delete(d.candidates, name)
		}
	}

	// Skip ordering if we have no new collector
	if !foundNew {
		return
	}

	// Pick prefered collector among detected ones
	var prefered string
	for name, _ := range d.detected {
		// First one
		if prefered == "" {
			prefered = name
			continue
		}
		// Highest priority first
		if CollectorPriorities[name] > CollectorPriorities[prefered] {
			prefered = name
			continue
		}
		// Alphabetic order to stay stable
		if CollectorPriorities[name] == CollectorPriorities[prefered] && strings.Compare(prefered, name) > 1 {
			prefered = name
			continue
		}
	}

	log.Infof("Using collector %s", prefered)
	d.preferedName = prefered
	d.preferedCollector = d.detected[prefered]
}
