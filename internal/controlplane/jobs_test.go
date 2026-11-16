package controlplane

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateCreateJob(t *testing.T) {
	valid:=createJobInput{TenantID:"tenant-a",SourceCommit:strings.Repeat("a",40),RequestedResources:json.RawMessage(`{}`)}
	if msg:=validateCreateJob(&valid);msg!=""{t.Fatalf("valid job rejected: %s",msg)}
	invalid:=createJobInput{TenantID:"tenant-a",SourceCommit:"not-a-commit"}
	if msg:=validateCreateJob(&invalid);msg==""{t.Fatal("invalid commit accepted")}
}
