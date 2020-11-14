package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSetEnvVar(t *testing.T) {
	k1 := "value1"
	k2 := "value2"
	os.Setenv("KEY1", "new_value")
	setEnvVar("KEY1", &k1)
	setEnvVar("KEY2", &k2)

	if k1 != "new_value" {
		t.Errorf("setEnvVar failed. expect new_value, but got %s", k1)
	}
	if k2 != "value2" {
		t.Errorf("setEnvVar failed. expect value2, but got %s", k2)
	}
}

func TestGetConfigFileName(t *testing.T) {
	tt := []struct {
		args     []string
		expectOK bool
		fname    string
	}{
		{
			args:     []string{},
			expectOK: true,
			fname:    "",
		},
		{
			args:     []string{"--config=test.yaml"},
			expectOK: true,
			fname:    "test.yaml",
		},
		{
			args:     []string{"-config=test.yaml"},
			expectOK: true,
			fname:    "test.yaml",
		},
		{
			args:     []string{"--config", "test.yaml"},
			expectOK: true,
			fname:    "test.yaml",
		},
		{
			args:     []string{"-config", "test.yaml"},
			expectOK: true,
			fname:    "test.yaml",
		},
		{
			args:     []string{"--test=test.yaml"},
			expectOK: true,
			fname:    "",
		},
		{
			args:     []string{"--config"},
			expectOK: false,
		},
		{
			args:     []string{"--config", "--test"},
			expectOK: false,
		},
	}

	for _, tc := range tt {
		f, err := getConfigFileName(tc.args)
		if tc.expectOK && err != nil {
			t.Errorf("getConfigFileName returns wrong response. input: %v, got %v, want nil", tc.args, err)
		}
		if !tc.expectOK && err == nil {
			t.Errorf("getConfigFileName returns wrong response. input: %v, got nil, but want not nil", tc.args)
		}
		if tc.expectOK && f != tc.fname {
			t.Errorf("getConfigFileName returns wrong file name. expect %s, but got %s", tc.fname, f)
		}
	}
}

func TestCheckLoginResDirStruct(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create tmp dir: %v", err)
		return
	}
	defer os.RemoveAll(dir)

	c := &GlobalConfig{
		UserLoginResourceDir: dir,
	}

	consentFile := filepath.Join(dir, "consent.html")
	errorFile := filepath.Join(dir, "error.html")
	indexFile := filepath.Join(dir, "index.html")
	deviceFile := filepath.Join(dir, "devicelogin.html")
	data := []byte("data")

	// Test no consent page
	// Error check only once
	if err := ioutil.WriteFile(errorFile, data, 0644); err != nil {
		t.Errorf("Failed to create error file: %v", err)
		return
	}
	ioutil.WriteFile(indexFile, data, 0644)
	ioutil.WriteFile(deviceFile, data, 0644)
	if err := c.setLoginResource(); err == nil {
		t.Errorf("CheckLoginResDirStruct returns nil, but expect is no consent page")
	}
	t.Logf("no consent page error: %v", err)
	// Error check only once
	if err := os.Remove(errorFile); err != nil {
		t.Errorf("Failed to delete error file: %v", err)
		return
	}
	os.Remove(indexFile)
	os.Remove(deviceFile)

	// Test no error page
	ioutil.WriteFile(consentFile, data, 0644)
	ioutil.WriteFile(indexFile, data, 0644)
	ioutil.WriteFile(deviceFile, data, 0644)
	if err := c.setLoginResource(); err == nil {
		t.Errorf("CheckLoginResDirStruct returns nil, but expect is no error page")
	}
	os.Remove(consentFile)
	os.Remove(indexFile)
	os.Remove(deviceFile)

	// Test no login page
	ioutil.WriteFile(consentFile, data, 0644)
	ioutil.WriteFile(errorFile, data, 0644)
	ioutil.WriteFile(deviceFile, data, 0644)
	if err := c.setLoginResource(); err == nil {
		t.Errorf("CheckLoginResDirStruct returns nil, but expect is no login page")
	}
	os.Remove(consentFile)
	os.Remove(errorFile)
	os.Remove(deviceFile)

	// Test no device login page
	ioutil.WriteFile(consentFile, data, 0644)
	ioutil.WriteFile(errorFile, data, 0644)
	ioutil.WriteFile(indexFile, data, 0644)
	if err := c.setLoginResource(); err == nil {
		t.Errorf("CheckLoginResDirStruct returns nil, but expect is no device login page")
	}
	os.Remove(consentFile)
	os.Remove(errorFile)
	os.Remove(indexFile)

	// Test ok
	ioutil.WriteFile(consentFile, data, 0644)
	ioutil.WriteFile(errorFile, data, 0644)
	ioutil.WriteFile(indexFile, data, 0644)
	ioutil.WriteFile(deviceFile, data, 0644)
	if err := c.setLoginResource(); err != nil {
		t.Errorf("CheckLoginResDirStruct returns error %v, but expect is nil", err)
	}
	os.Remove(consentFile)
	os.Remove(errorFile)
	os.Remove(indexFile)
	os.Remove(deviceFile)
}
