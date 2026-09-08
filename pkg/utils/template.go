package utils

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/redhat-developer/rhdh-operator/pkg/platform"
)

// TemplateData provides values for Go template substitution in config files.
type TemplateData struct {
	Backstage BackstageInfo
	OpenShift OpenShiftInfo
	Platform  platform.Platform
}

// BackstageInfo contains Backstage CR fields available for templating in config files.
type BackstageInfo struct {
	Name      string
	Namespace string
}

// OpenShiftInfo contains OpenShift-specific values available for templating.
type OpenShiftInfo struct {
	IngressDomain string
}

// templateData holds the current template data for YAML processing.
var templateData *TemplateData

// SetTemplateData sets the template data for YAML file processing.
// Call this once before reading config files.
func SetTemplateData(name, namespace, openShiftIngressDomain string, detectedPlatform platform.Platform) {
	templateData = &TemplateData{
		Backstage: BackstageInfo{
			Name:      name,
			Namespace: namespace,
		},
		OpenShift: OpenShiftInfo{
			IngressDomain: openShiftIngressDomain,
		},
		Platform: detectedPlatform,
	}
}

// ApplyTemplate applies Go template substitution to content if templateData is set
// and the content contains one of our supported template variables.
// Returns content unchanged if no template data has been set or no template variables found.
func ApplyTemplate(content []byte) ([]byte, error) {
	if templateData == nil {
		return content, nil
	}
	// Only parse as a template if a template action references one of our
	// supported data groups. This avoids parsing application-owned patterns such
	// as {{message}}, while supporting control actions such as
	// {{if eq .Platform.Extension "ocp"}}.
	if !containsSupportedTemplateField(string(content)) {
		return content, nil
	}
	tmpl, err := template.New("config").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.Bytes(), nil
}

func containsSupportedTemplateField(content string) bool {
	for {
		start := strings.Index(content, "{{")
		if start == -1 {
			return false
		}
		content = content[start+2:]
		end := strings.Index(content, "}}")
		if end == -1 {
			return false
		}
		action := content[:end]
		if strings.Contains(action, ".Backstage.") ||
			strings.Contains(action, ".OpenShift.") ||
			strings.Contains(action, ".Platform.") {
			return true
		}
		content = content[end+2:]
	}
}
