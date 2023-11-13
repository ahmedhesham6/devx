package drivers

import (
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/stakpak/devx/pkg/stack"
	"github.com/stakpak/devx/pkg/stackbuilder"
	"github.com/stakpak/devx/pkg/utils"
)

type ComposeDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *ComposeDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "compose"
}

func (d *ComposeDriver) ApplyAll(stack *stack.Stack, stdout bool) error {

	composeFile := stack.GetContext().CompileString("_")
	foundResources := false

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				foundResources = true
				composeFile = composeFile.Fill(resourceIter.Value())
			}
		}
	}

	if !foundResources {
		return nil
	}

	composeFile, err := utils.RemoveMeta(composeFile)
	if err != nil {
		return err
	}
	data, err := yaml.Encode(composeFile)
	if err != nil {
		return err
	}

	if stdout {
		_, err := os.Stdout.Write(data)
		return err
	}

	if _, err := os.Stat(d.Config.Output.Dir); os.IsNotExist(err) {
		if err := os.MkdirAll(d.Config.Output.Dir, 0700); err != nil {
			return err
		}
	}
	filePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return err
	}

	log.Infof("[compose] applied resources to \"%s\"", filePath)

	return nil
}
