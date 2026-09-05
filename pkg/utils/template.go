package utils

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateData provides values for Go template substitution in config files.
type TemplateData struct {
	Backstage BackstageInfo
	OpenShift OpenShiftInfo
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
func SetTemplateData(name, namespace, openShiftIngressDomain string) {
	templateData = &TemplateData{
		Backstage: BackstageInfo{
			Name:      name,
			Namespace: namespace,
		},
		OpenShift: OpenShiftInfo{
			IngressDomain: openShiftIngressDomain,
		},
	}
}

// ApplyTemplate applies Go template substitution to content if templateData is set
// and the content contains our template variables ({{.Backstage.}}).
// Returns content unchanged if no template data has been set or no template variables found.
func ApplyTemplate(content []byte) ([]byte, error) {
	if templateData == nil {
		return content, nil
	}
	// Only parse as template if our specific variables are present
	// This avoids parsing errors from other {{...}} patterns in config files
	if !strings.Contains(string(content), "{{.Backstage.") &&
		!strings.Contains(string(content), "{{.OpenShift.") {
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
