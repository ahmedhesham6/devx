package stackbuilder

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
)

type Environments = map[string]*StackBuilder

type StackBuilder struct {
	// AdditionalComponents *cue.Value
	Flows []*Flow
}

func NewEnvironments(value cue.Value) (Environments, error) {
	environments := map[string]*StackBuilder{}

	envIter, err := value.Fields()
	if err != nil {
		return environments, err
	}

	for envIter.Next() {
		name := utils.GetLastPathFragement(envIter.Value())
		environments[name], err = NewStackBuilder(envIter.Value())
		if err != nil {
			return environments, err
		}
	}

	return environments, nil
}

func NewStackBuilder(value cue.Value) (*StackBuilder, error) {
	flows := value.LookupPath(cue.ParsePath("flows"))
	if flows.Err() != nil {
		return nil, flows.Err()
	}

	stackBuilder := StackBuilder{
		// AdditionalComponents: nil,
		Flows: make([]*Flow, 0),
	}
	flowIter, _ := flows.List()
	for flowIter.Next() {
		flow, err := NewFlow(flowIter.Value())
		if err != nil {
			return nil, err
		}
		stackBuilder.Flows = append(stackBuilder.Flows, flow)
	}

	return &stackBuilder, nil
}

func (sb *StackBuilder) TransformStack(stack *stack.Stack) error {
	orderedTasks := stack.GetTasks()
	for _, componentId := range orderedTasks {
		for _, flow := range sb.Flows {
			err := flow.Run(stack, componentId)
			if err != nil {
				return err
			}
			if !stack.HasConcreteResourceDrivers(componentId) {
				return fmt.Errorf(
					"Component %s resources do not have concrete drivers",
					componentId,
				)
			}
		}
		if !stack.IsConcreteComponent(componentId) {
			// find all errors
			errors := []string{}
			c, _ := stack.GetComponent(componentId)

			c.Walk(func(_ cue.Value) bool { return true }, func(value cue.Value) {
				if value.Err() != nil {
					errors = append(errors, value.Err().Error())
				}

				if value.Validate() != nil {
					errors = append(errors, fmt.Sprintf("%s: %s", value.Path(), value.Validate().Error()))
				}
			})

			return fmt.Errorf("component %s is not concrete after transformation:\n  %s", componentId, strings.Join(errors, "\n  "))
		}
	}
	return nil
}
