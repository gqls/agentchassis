package main

import (
	"fmt"
	"time"
)

// Create a test that spawns agents and uses them
func TestWebsiteBuilderWithSpawning() {
	// Create workflow that includes agent spawning
	workflow := models.WorkflowPlan{
		StartStep: "plan_team",
		Steps: map[string]models.Step{
			"plan_team": {
				Action: "plan_agent_team",
				Config: map[string]interface{}{
					"task_type": "website_builder",
					"requirements": map[string]interface{}{
						"capabilities": []string{"html", "css", "design"},
					},
				},
				NextStep: "spawn_team",
			},
			"spawn_team": {
				Action: "spawn_group",
				Config: map[string]interface{}{
					"group_type": "website-builder",
				},
				NextStep: "execute_task",
			},
			"execute_task": {
				Action: "fan_out",
				SubTasks: []models.SubTask{
					{StepName: "design", Topic: "system.agent.designer.process"},
					{StepName: "develop", Topic: "system.agent.developer.process"},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}
}
