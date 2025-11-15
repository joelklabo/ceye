package config

import (
	"os"
	"testing"
)

func TestLoadConfigFromYAML(t *testing.T) {
	data := []byte(`providers:
  - type: github
    repos:
      - owner: octocat
        repo: Hello-World
        workflows: ["CI", "Deploy"]
  - type: azure
    org: myorg
    project: myproject
    pipelines: [42, 43]
`)
	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp config: %v", err)
	}

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(cfg.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cfg.Providers))
	}

	githubCfg := cfg.Providers[0]
	if githubCfg.Type != "github" {
		t.Fatalf("expected first provider type github, got %s", githubCfg.Type)
	}
	if len(githubCfg.Repos) != 1 || githubCfg.Repos[0].Owner != "octocat" || githubCfg.Repos[0].Repo != "Hello-World" {
		t.Fatalf("unexpected github repo config: %+v", githubCfg.Repos)
	}

	azureCfg := cfg.Providers[1]
	if azureCfg.Type != "azure" {
		t.Fatalf("expected second provider type azure, got %s", azureCfg.Type)
	}
	if azureCfg.Org != "myorg" || azureCfg.Project != "myproject" {
		t.Fatalf("unexpected azure config: %+v", azureCfg)
	}
	expectedPipelines := []int{42, 43}
	if len(azureCfg.Pipelines) != len(expectedPipelines) {
		t.Fatalf("unexpected number of pipelines: %v", azureCfg.Pipelines)
	}
	for i, v := range expectedPipelines {
		if azureCfg.Pipelines[i] != v {
			t.Fatalf("unexpected pipeline id at %d: got %d want %d", i, azureCfg.Pipelines[i], v)
		}
	}
}
