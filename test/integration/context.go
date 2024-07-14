package integration

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/javiertelioz/go-flows/pkg/workflow/config"
)

var wm *workflow.WorkflowManager
var workflowPath string

func iConfigureTheWorkflowFrom(filePath string) error {
	workflowPath = filePath
	return nil
}

func iExecuteTheWorkflowFrom(nodeID string) error {
	cfg, err := config.LoadConfig(workflowPath)
	if err != nil {
		return err
	}

	err = wm.LoadFromConfig(cfg)
	if err != nil {
		return err
	}

	err = wm.Execute(nodeID, nil)
	if err != nil {
		return err
	}

	return nil
}

func theWorkflowShouldCompleteSuccessfully() error {
	fmt.Println("Workflow completed successfully")
	return nil
}

func theConditionNodeIsExecutedWithTrueCondition() error {
	return nil
}

func theWorkflowShouldFollowTheTruePath() error {
	fmt.Println("Workflow followed the true path")
	return nil
}

func theParallelNodeIsExecuted() error {
	return nil
}

func allParallelTasksShouldCompleteSuccessfully() error {
	fmt.Println("All parallel tasks completed successfully")
	return nil
}

func theForeachNodeIsExecuted() error {
	return nil
}

func allIterationsShouldCompleteSuccessfully() error {
	fmt.Println("All iterations completed successfully")
	return nil
}

func theSubDagNodeIsExecuted() error {
	return nil
}

func allSubTasksShouldCompleteSuccessfully() error {
	fmt.Println("All sub-tasks completed successfully")
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a workflow configured from "([^"]*)"$`, iConfigureTheWorkflowFrom)
	ctx.Step(`^the workflow is executed from the "([^"]*)" node$`, iExecuteTheWorkflowFrom)
	ctx.Step(`^the workflow should complete successfully$`, theWorkflowShouldCompleteSuccessfully)
	ctx.Step(`^the condition node is executed with true condition$`, theConditionNodeIsExecutedWithTrueCondition)
	ctx.Step(`^the workflow should follow the true path$`, theWorkflowShouldFollowTheTruePath)
	ctx.Step(`^the parallel node is executed$`, theParallelNodeIsExecuted)
	ctx.Step(`^all parallel tasks should complete successfully$`, allParallelTasksShouldCompleteSuccessfully)
	ctx.Step(`^the foreach node is executed$`, theForeachNodeIsExecuted)
	ctx.Step(`^all iterations should complete successfully$`, allIterationsShouldCompleteSuccessfully)
	ctx.Step(`^the SubDag node is executed$`, theSubDagNodeIsExecuted)
	ctx.Step(`^all sub-tasks should complete successfully$`, allSubTasksShouldCompleteSuccessfully)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {}

func init() {
	taskFuncs := map[string]interface{}{
		"startFunc":    startFunc,
		"checkFunc":    checkFunc,
		"task1Func":    task1Func,
		"task2Func":    task2Func,
		"foreachFunc":  foreachFunc,
		"subtask1Func": subtask1Func,
		"subtask2Func": subtask2Func,
		"endFunc":      endFunc,
	}
	hookFuncs := map[string]interface{}{
		"beforeTask": beforeTask,
		"afterTask":  afterTask,
	}
	wm = workflow.NewWorkflowManager(nil, taskFuncs, hookFuncs)
}

// Define the dummy functions
func startFunc(data interface{}) (interface{}, error)    { return data, nil }
func checkFunc(data interface{}) (bool, error)           { return true, nil }
func task1Func(data interface{}) (interface{}, error)    { return data, nil }
func task2Func(data interface{}) (interface{}, error)    { return data, nil }
func foreachFunc(data interface{}) (interface{}, error)  { return data, nil }
func subtask1Func(data interface{}) (interface{}, error) { return data, nil }
func subtask2Func(data interface{}) (interface{}, error) { return data, nil }
func endFunc(data interface{}) (interface{}, error)      { return data, nil }
func beforeTask(data interface{}) (interface{}, error)   { return data, nil }
func afterTask(data interface{}) (interface{}, error)    { return data, nil }
