package installer

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"gopkg.in/yaml.v3"
)

// WorkflowStepModel identifies one subagent step of a specialist workflow and
// the model it runs on. Model is the source default from workflow.yaml; an
// empty Model means the step has no default and falls back to the global model.
type WorkflowStepModel struct {
	Specialist string
	Step       string
	Model      string
}

// Key returns the selection-map key for this step: "{specialist}/{step}".
func (w WorkflowStepModel) Key() string {
	return w.Specialist + "/" + w.Step
}

// workflowModelStep is the minimal step shape needed to list model steps.
type workflowModelStep struct {
	Name      string `yaml:"name"`
	Execution string `yaml:"execution"`
	Model     string `yaml:"model"`
}

// workflowModelFile is the minimal workflow.yaml shape needed to list model steps.
type workflowModelFile struct {
	Specialist string              `yaml:"specialist"`
	Steps      []workflowModelStep `yaml:"steps"`
}

// WorkflowModelSteps walks skillsFS for */workflow.yaml files and returns
// every `execution: subagent` step with its specialist and source-default
// model. Inline steps never launch their own agent, so they have no model to
// pick. Results are ordered by specialist directory name, then step order.
func WorkflowModelSteps(skillsFS fs.FS) ([]WorkflowStepModel, error) {
	entries, err := fs.ReadDir(skillsFS, ".")
	if err != nil {
		return nil, fmt.Errorf("read skill root: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var steps []WorkflowStepModel
	for _, dir := range names {
		wfPath := path.Join(dir, "workflow.yaml")
		data, readErr := fs.ReadFile(skillsFS, wfPath)
		if readErr != nil {
			continue // specialist without workflow.yaml (e.g. asdt-init, asdt-shared)
		}

		var wf workflowModelFile
		if err := yaml.Unmarshal(data, &wf); err != nil {
			return nil, fmt.Errorf("parse %s: %w", wfPath, err)
		}

		for _, s := range wf.Steps {
			if s.Execution != "subagent" {
				continue
			}
			steps = append(steps, WorkflowStepModel{
				Specialist: wf.Specialist,
				Step:       s.Name,
				Model:      s.Model,
			})
		}
	}

	return steps, nil
}

// InjectModels returns workflow.yaml content with each subagent step's
// `model:` field set from models (keyed "{specialist}/{step}"). Steps without
// a selection are left untouched. The YAML is round-tripped through yaml.Node
// so comments and field order survive; a nil/empty models map returns content
// unchanged without parsing.
func InjectModels(content []byte, models map[string]string) ([]byte, error) {
	if len(models) == 0 {
		return content, nil
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse workflow.yaml: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return content, nil
	}

	doc := root.Content[0]
	specialist := mappingValue(doc, "specialist")
	stepsNode := mappingNode(doc, "steps")
	if specialist == nil || stepsNode == nil || stepsNode.Kind != yaml.SequenceNode {
		return content, nil
	}

	changed := false
	for _, step := range stepsNode.Content {
		if step.Kind != yaml.MappingNode {
			continue
		}
		name := mappingValue(step, "name")
		if name == nil {
			continue
		}
		model, ok := models[specialist.Value+"/"+name.Value]
		if !ok {
			continue
		}
		// A selection equal to the existing value is a no-op — a user who
		// cycled and came back to the default must not trigger a re-encode.
		if existing := mappingValue(step, "model"); existing != nil && existing.Value == model {
			continue
		}
		setMappingValue(step, "model", model)
		changed = true
	}

	if !changed {
		return content, nil
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encode workflow.yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close encoder: %w", err)
	}
	return buf.Bytes(), nil
}

// RemoveModels returns workflow.yaml content with the `model:` field stripped
// from every subagent step, so each step inherits the model the assistant
// already has defined (the Chameleon preset). The YAML is round-tripped through
// yaml.Node so comments and field order survive; when no step carries a model
// field the content is returned unchanged without re-encoding.
func RemoveModels(content []byte) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse workflow.yaml: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return content, nil
	}

	doc := root.Content[0]
	stepsNode := mappingNode(doc, "steps")
	if stepsNode == nil || stepsNode.Kind != yaml.SequenceNode {
		return content, nil
	}

	changed := false
	for _, step := range stepsNode.Content {
		if step.Kind != yaml.MappingNode {
			continue
		}
		if deleteMappingKey(step, "model") {
			changed = true
		}
	}

	if !changed {
		return content, nil
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encode workflow.yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close encoder: %w", err)
	}
	return buf.Bytes(), nil
}

// deleteMappingKey removes the key/value pair for key from a mapping node,
// dropping the two adjacent Content slots that hold them. It reports whether a
// pair was removed.
func deleteMappingKey(mapping *yaml.Node, key string) bool {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return true
		}
	}
	return false
}

// mappingValue returns the value node for key in a mapping node, or nil.
func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// mappingNode is mappingValue under a clearer name for nested-structure access.
func mappingNode(mapping *yaml.Node, key string) *yaml.Node {
	return mappingValue(mapping, key)
}

// setMappingValue updates the value for key in a mapping node, inserting the
// pair right after "name" (or appending) when the key is absent.
func setMappingValue(mapping *yaml.Node, key, value string) {
	if existing := mappingValue(mapping, key); existing != nil {
		existing.SetString(value)
		return
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valNode := &yaml.Node{Kind: yaml.ScalarNode, Value: value}

	insertAt := len(mapping.Content)
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == "name" {
			insertAt = i + 2
			break
		}
	}

	mapping.Content = append(mapping.Content, nil, nil)
	copy(mapping.Content[insertAt+2:], mapping.Content[insertAt:])
	mapping.Content[insertAt] = keyNode
	mapping.Content[insertAt+1] = valNode
}
