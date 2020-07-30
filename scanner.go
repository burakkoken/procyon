package procyon

import (
	"fmt"
	core "github.com/procyon-projects/procyon-core"
	peas "github.com/procyon-projects/procyon-peas"
)

type componentScanner struct {
}

func newComponentScanner() componentScanner {
	return componentScanner{}
}

func (scanner componentScanner) scan(logger core.Logger) (int, error) {
	processors, err := scanner.getProcessorInstances()
	if err != nil {
		return -1, nil
	}
	var componentCount = 0
	componentMap := core.GetComponentTypeMap()
	for componentName := range componentMap {
		component := componentMap[componentName]
		logger.Trace(fmt.Sprintf("Component : %s", componentName))
		for _, processorInstance := range processors {
			if processor, ok := processorInstance.(core.ComponentProcessor); ok {
				if processor.SupportsComponent(component) {
					err = processor.ProcessComponent(component)
					if err != nil {
						return -1, err
					}
				}
			}
		}
		componentCount++
	}
	return componentCount, err
}

func (scanner componentScanner) getProcessorInstances() ([]interface{}, error) {
	componentProcessors := core.GetComponentProcessorMap()
	var instances []interface{}
	for componentName := range componentProcessors {
		processorType := componentProcessors[componentName]
		instance, err := peas.CreateInstance(processorType, []interface{}{})
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}
