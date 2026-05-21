package templates

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Template represents a command template with metadata
type Template struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
}

// TemplateData represents the structure of the templates YAML file
type TemplateData struct {
	Templates []Template `yaml:"templates"`
}

// Loader handles loading command templates from YAML files
type Loader struct {
	templatePath string
}

// NewLoader creates a new template loader
func NewLoader(templatePath string) *Loader {
	return &Loader{
		templatePath: templatePath,
	}
}

// Load loads templates from the configured file
func (l *Loader) Load() ([]Template, error) {
	// Check if file exists
	if _, err := os.Stat(l.templatePath); os.IsNotExist(err) {
		// Create default templates file
		err := l.createDefaultTemplates()
		if err != nil {
			return nil, err
		}
	}

	// Read the file
	data, err := os.ReadFile(l.templatePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var templateData TemplateData
	err = yaml.Unmarshal(data, &templateData)
	if err != nil {
		return nil, err
	}

	// Sort templates by category, then by name
	sort.Slice(templateData.Templates, func(i, j int) bool {
		if templateData.Templates[i].Category != templateData.Templates[j].Category {
			return templateData.Templates[i].Category < templateData.Templates[j].Category
		}
		return templateData.Templates[i].Name < templateData.Templates[j].Name
	})

	return templateData.Templates, nil
}

func (l *Loader) Add(command string) (Template, error) {
    existing, err := l.Load()
    if err != nil {
        return Template{}, err
    }

    for _, t := range existing {
        if t.Command == command {
            return t, nil // уже есть, не добавлять
        }
    }

    t := Template{
        Name:     truncateTemplateName(command, 40),
        Command:  command,
        Category: detectCategory(command),
    }

    data := TemplateData{Templates: append(existing, t)}
    return t, l.save(data)
}

func (l *Loader) Delete(command string) error {
    existing, err := l.Load()
    if err != nil {
        return err
    }

    filtered := existing[:0]
    for _, t := range existing {
        if t.Command != command {
            filtered = append(filtered, t)
        }
    }

    return l.save(TemplateData{Templates: filtered})
}

func (l *Loader) save(data TemplateData) error {
    raw, err := yaml.Marshal(data)
    if err != nil {
        return err
    }
    return os.WriteFile(l.templatePath, raw, 0644)
}

func truncateTemplateName(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max-3] + "..."
}

func detectCategory(command string) string {
    prefixes := map[string]string{
        "git":      "git",
        "docker":   "docker",
        "kubectl":  "k8s",
        "ssh":      "ssh",
        "make":     "build",
        "go":       "go",
        "npm":      "node",
        "yarn":     "node",
        "systemctl":"system",
        "brew":     "system",
        "apt":      "system",
    }
    parts := strings.Fields(command)
    if len(parts) == 0 {
        return "other"
    }
    if cat, ok := prefixes[parts[0]]; ok {
        return cat
    }
    return "other"
}

// createDefaultTemplates creates a default templates file
func (l *Loader) createDefaultTemplates() error {
	defaultTemplates := TemplateData{
		Templates: []Template{
			{
				Name:        "Git status",
				Command:     "git status",
				Description: "Show working tree status",
				Category:    "git",
			},
			{
				Name:        "Git log oneline",
				Command:     "git log --oneline -10",
				Description: "Show last 10 commits in one line",
				Category:    "git",
			},
			{
				Name:        "Git branch",
				Command:     "git branch -a",
				Description: "List all branches",
				Category:    "git",
			},
			{
				Name:        "Git diff",
				Command:     "git diff",
				Description: "Show changes in working directory",
				Category:    "git",
			},
			{
				Name:        "Docker ps",
				Command:     "docker ps -a",
				Description: "List all containers",
				Category:    "docker",
			},
			{
				Name:        "Docker images",
				Command:     "docker images",
				Description: "List all images",
				Category:    "docker",
			},
			{
				Name:        "Docker logs",
				Command:     "docker logs -f",
				Description: "Follow container logs",
				Category:    "docker",
			},
			{
				Name:        "Disk usage",
				Command:     "df -h",
				Description: "Show disk space usage",
				Category:    "system",
			},
			{
				Name:        "Memory usage",
				Command:     "free -h",
				Description: "Show memory usage",
				Category:    "system",
			},
			{
				Name:        "Process tree",
				Command:     "pstree",
				Description: "Display running processes as tree",
				Category:    "system",
			},
			{
				Name:        "Top processes",
				Command:     "top",
				Description: "Display running processes",
				Category:    "system",
			},
			{
				Name:        "Network connections",
				Command:     "netstat -tulpn",
				Description: "Show network connections",
				Category:    "network",
			},
			{
				Name:        "Ping test",
				Command:     "ping -c 4 google.com",
				Description: "Test network connectivity",
				Category:    "network",
			},
			{
				Name:        "Find files",
				Command:     "find . -name",
				Description: "Find files by name",
				Category:    "files",
			},
			{
				Name:        "Archive create",
				Command:     "tar -czf archive.tar.gz",
				Description: "Create compressed archive",
				Category:    "files",
			},
		},
	}

	// Create directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(l.templatePath), 0755)
	if err != nil {
		return err
	}

	// Write templates to file
	data, err := yaml.Marshal(defaultTemplates)
	if err != nil {
		return err
	}

	return os.WriteFile(l.templatePath, data, 0644)
}

// GetByCategory returns templates grouped by category
func GetByCategory(templates []Template) map[string][]Template {
	categories := make(map[string][]Template)

	for _, template := range templates {
		category := template.Category
		if category == "" {
			category = "other"
		}

		if _, exists := categories[category]; !exists {
			categories[category] = make([]Template, 0)
		}

		categories[category] = append(categories[category], template)
	}

	return categories
}

// Search finds templates matching the query
func Search(templates []Template, query string) []Template {
	if query == "" {
		return templates
	}

	var results []Template
	queryLower := strings.ToLower(query)

	for _, template := range templates {
		if strings.Contains(strings.ToLower(template.Name), queryLower) ||
			strings.Contains(strings.ToLower(template.Command), queryLower) ||
			strings.Contains(strings.ToLower(template.Description), queryLower) ||
			strings.Contains(strings.ToLower(template.Category), queryLower) {
			results = append(results, template)
		}
	}

	return results
}
