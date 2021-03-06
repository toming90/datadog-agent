// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build docker

package collectors

import (
	"time"

	"github.com/DataDog/datadog-agent/pkg/tagger/utils"
	"github.com/DataDog/datadog-agent/pkg/util/docker"
	ecsutil "github.com/DataDog/datadog-agent/pkg/util/ecs"
)

func (c *ECSCollector) parseTasks(tasks_list ecsutil.TasksV1Response) ([]*TagInfo, error) {
	var output []*TagInfo
	now := time.Now()
	for _, task := range tasks_list.Tasks {
		// We only want to collect tasks without a STOPPED status.
		if task.KnownStatus == "STOPPED" {
			continue
		}
		for _, container := range task.Containers {
			// We only want to collect the tags from new containers.
			if c.expire.Update(container.DockerID, now) {
				tags := utils.NewTagList()
				tags.AddLow("task_version", task.Version)
				tags.AddLow("task_name", task.Family)

				low, high := tags.Compute()

				info := &TagInfo{
					Source:       ecsCollectorName,
					Entity:       docker.ContainerIDToEntityName(container.DockerID),
					HighCardTags: high,
					LowCardTags:  low,
				}
				output = append(output, info)
			}
		}
	}
	return output, nil
}
