package db

import (
	"testing"
	"time"

	"github.com/sh-miyoshi/hekate/pkg/db/memory"
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"github.com/sh-miyoshi/hekate/pkg/errors"
)

func TestProjectAdd(t *testing.T) {
	mgr := &Manager{
		client:      memory.NewClientHandler(),
		project:     memory.NewProjectHandler(),
		transaction: memory.NewTransactionManager(),
	}

	prjInfo := &model.ProjectInfo{
		Name:      "test-project",
		CreatedAt: time.Now(),
		TokenConfig: &model.TokenConfig{
			AccessTokenLifeSpan:  1,
			RefreshTokenLifeSpan: 1,
			SigningAlgorithm:     "RS256",
		},
	}

	// Test Correct Project
	if err := mgr.ProjectAdd(prjInfo); err != nil {
		t.Errorf("Failed to add correct project: %v", err)
	}

	// Check portal client exists
	clis, _ := mgr.client.GetList(prjInfo.Name, &model.ClientFilter{ID: "portal"})
	if len(clis) != 1 {
		t.Errorf("Failed to register portal client, expect 1 client, but got %d", len(clis))
	}

	// Test Duplicate Project Name
	err := mgr.ProjectAdd(prjInfo)
	if !errors.Contains(err, model.ErrProjectAlreadyExists) {
		t.Errorf("Expect error is %v, but got %v", model.ErrProjectAlreadyExists, err)
	}
}
