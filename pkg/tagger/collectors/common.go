// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package collectors

import (
	"strings"

	"github.com/DataDog/datadog-agent/pkg/util/tmplvar"
)

func resolveTag(tmpl, label string) string {
	vars := tmplvar.ParseString(tmpl)
	tagName := tmpl
	for old, v := range vars {
		if string(v.Name) != "label" {
			tagName = strings.Replace(tagName, old, "", -1)
			continue
		}
		tagName = strings.Replace(tagName, old, label, -1)
	}
	return tagName
}
