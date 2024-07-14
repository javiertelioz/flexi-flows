Feature: Workflow execution

  Scenario: Execute full workflow with all node types
    Given a workflow configured from "config/workflow.json"
    When the workflow is executed from the "start" node
    Then the workflow should complete successfully

  Scenario: Conditional node checks
    Given a workflow configured from "config/workflow.json"
    When the condition node is executed with true condition
    Then the workflow should follow the true path

  Scenario: Parallel task execution
    Given a workflow configured from "config/workflow.json"
    When the parallel node is executed
    Then all parallel tasks should complete successfully

  Scenario: Foreach node iteration
    Given a workflow configured from "config/workflow.json"
    When the foreach node is executed
    Then all iterations should complete successfully

  Scenario: SubDag execution
    Given a workflow configured from "config/workflow.json"
    When the SubDag node is executed
    Then all sub-tasks should complete successfully
