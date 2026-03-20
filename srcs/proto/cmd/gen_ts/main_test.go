package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempProto(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.proto")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write proto: %v", err)
	}
	return path
}

func TestSnakeToCamel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"id", "id"},
		{"organization_id", "organizationId"},
		{"from_agent", "fromAgent"},
		{"occurred_at_unix", "occurredAtUnix"},
		{"base_prompt", "basePrompt"},
		{"is_human", "isHuman"},
		{"total_cost_usd", "totalCostUsd"},
	}
	for _, c := range cases {
		if got := snakeToCamel(c.in); got != c.want {
			t.Errorf("snakeToCamel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPkgToPrefix(t *testing.T) {
	cases := []struct{ pkg, want string }{
		{"ohc.common", "Common"},
		{"ohc.organization", "Organization"},
		{"ohc.billing", "Billing"},
		{"ohc.orchestration", "Orchestration"},
		{"ohc.api.v1", "Api"},
		{"ohc.api.v2alpha", "Api"},
		{"ohc.agent", "Agent"},
		{"ohc.skills", "Skills"},
		{"ohc", "Proto"},
		{"", "Proto"},
	}
	for _, c := range cases {
		if got := pkgToPrefix(c.pkg); got != c.want {
			t.Errorf("pkgToPrefix(%q) = %q, want %q", c.pkg, got, c.want)
		}
	}
}

func TestQualifiedToTS(t *testing.T) {
	allTypes := map[string]string{
		"Role": "CommonRole",
		"ohc.common.Role": "CommonRole",
	}

	cases := []struct{ in, want string }{
		{"Role", "CommonRole"},
		{".Role", "CommonRole"},
		{"ohc.common.Role", "CommonRole"},
		{".ohc.common.Role", "CommonRole"},
		{"Unknown", "Unknown"},
		{"ohc.api.Unknown", "Unknown"},
	}
	for _, c := range cases {
		got := qualifiedToTS(c.in, allTypes)
		if got != c.want {
			t.Errorf("qualifiedToTS(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestProtoTypeToTS(t *testing.T) {
	cases := []struct{ in, want string }{
		{"string", "string"},
		{"bool", "boolean"},
		{"int32", "number"},
		{"int64", "number"},
		{"uint32", "number"},
		{"uint64", "number"},
		{"sint32", "number"},
		{"sint64", "number"},
		{"fixed32", "number"},
		{"fixed64", "number"},
		{"sfixed32", "number"},
		{"sfixed64", "number"},
		{"float", "number"},
		{"double", "number"},
		{"bytes", "string"},
		{"MyMessage", ""},
		{".ohc.common.Role", ""},
	}
	for _, c := range cases {
		got := protoTypeToTS(c.in)
		if got != c.want {
			t.Errorf("protoTypeToTS(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestStripComment(t *testing.T) {
	cases := []struct{ in, want string }{
		{"message Foo {", "message Foo {"},
		{"message Foo { // hello", "message Foo {"},
		{"message Foo { /* hello */", "message Foo {"},
		{"  string name = 1;  ", "string name = 1;"},
	}
	for _, c := range cases {
		got := stripComment(c.in)
		if got != c.want {
			t.Errorf("stripComment(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseEnums(t *testing.T) {
	proto := `
edition = "2024";
package ohc.common;
option go_package = "github.com/example;commonpb";

enum Role {
  ROLE_UNSPECIFIED = 0;
  CEO = 1;
  PRODUCT_MANAGER = 2;
}
`
	path := writeTempProto(t, proto)
	pf := parseProto(path)
	if pf.pkg != "ohc.common" {
		t.Errorf("expected package ohc.common, got %q", pf.pkg)
	}
	if len(pf.enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(pf.enums))
	}
	if pf.enums[0].name != "Role" {
		t.Errorf("expected enum name Role, got %q", pf.enums[0].name)
	}
	if len(pf.enums[0].values) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(pf.enums[0].values))
	}
}

func TestParseMessages(t *testing.T) {
	proto := `
syntax = "proto3";
package ohc.organization;

message TeamMember {
  string id = 1;
  string organization_id = 2;
  string name = 3;
  bool is_human = 4;
}
`
	path := writeTempProto(t, proto)
	pf := parseProto(path)
	if len(pf.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(pf.messages))
	}
	msg := pf.messages[0]
	if msg.name != "TeamMember" {
		t.Errorf("expected name TeamMember, got %q", msg.name)
	}
	if len(msg.fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(msg.fields))
	}
}

func TestGenerateOutput(t *testing.T) {
	proto := `
syntax = "proto3";
package ohc.billing;

enum AlertLevel {
  NONE = 0;
  LOW = 1;
  HIGH = 2;
}

message CostSummary {
  string organization_id = 1;
  double total_cost_usd = 2;
  int64 total_tokens = 3;
  repeated AgentCostSummary agents = 4;
}

message AgentCostSummary {
  string agent_id = 1;
  double cost_usd = 2;
  int64 token_used = 3;
}
`
	path := writeTempProto(t, proto)
	pf := parseProto(path)

	// Capture output.
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	generate([]protoFile{pf})
	w.Close()
	os.Stdout = old

	outBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	out := string(outBytes)

	checks := []string{
		"export type BillingAlertLevel =",
		`"NONE"`,
		`"LOW"`,
		`"HIGH"`,
		"export interface BillingCostSummary {",
		"organizationId?: string;",
		"totalCostUsd?: number;",
		"totalTokens?: number;",
		"agents?: BillingAgentCostSummary[];",
		"export interface BillingAgentCostSummary {",
		"agentId?: string;",
		"costUsd?: number;",
		"tokenUsed?: number;",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nFull output:\n%s", want, out)
		}
	}
}

func TestGenerateHeader(t *testing.T) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	generate(nil)
	w.Close()
	os.Stdout = old

	outBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	out := string(outBytes)

	if !strings.Contains(out, "GENERATED FILE") {
		t.Errorf("expected GENERATED FILE header, got: %s", out)
	}
}

func TestParseRepeatedFields(t *testing.T) {
	proto := `
syntax = "proto3";
package ohc.orchestration;

message MeetingRoom {
  string id = 1;
  repeated string participants = 2;
  repeated Message transcript = 3;
}

message Message {
  string id = 1;
  string content = 2;
}
`
	path := writeTempProto(t, proto)
	pf := parseProto(path)

	// Find MeetingRoom message.
	var mr *messageDef
	for i := range pf.messages {
		if pf.messages[i].name == "MeetingRoom" {
			mr = &pf.messages[i]
			break
		}
	}
	if mr == nil {
		t.Fatal("expected MeetingRoom message")
	}

	var participantsField, transcriptField *fieldDef
	for i := range mr.fields {
		switch mr.fields[i].name {
		case "participants":
			participantsField = &mr.fields[i]
		case "transcript":
			transcriptField = &mr.fields[i]
		}
	}
	if participantsField == nil || !participantsField.repeated {
		t.Error("expected participants to be repeated")
	}
	if transcriptField == nil || !transcriptField.repeated {
		t.Error("expected transcript to be repeated")
	}
}

func TestExecute_MissingArgs(t *testing.T) {
	// Let's stub out the exit behavior for the test
	originalOsExit := osExit
	t.Cleanup(func() { osExit = originalOsExit })
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}

	execute([]string{})
	if !exitCalled {
		t.Fatal("expected osExit to be called")
	}
}

func TestMainFunc(t *testing.T) {
	originalOsArgs := os.Args
	t.Cleanup(func() { os.Args = originalOsArgs })

	proto := `
		package ohc.main;
		message MainMsg {
			string item = 1;
		}
	`
	f := writeTempProto(t, proto)

	os.Args = []string{"gen_ts", f}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "export interface MainMainMsg") {
		t.Errorf("expected generated output to contain MainMainMsg, got:\n%s", output)
	}
}

func TestExecute_ValidArgs(t *testing.T) {
	proto := `
		package ohc.main;
		message MainMsg {
			string item = 1;
		}
	`
	f := writeTempProto(t, proto)

	// capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	execute([]string{f})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "export interface MainMainMsg") {
		t.Errorf("expected generated output to contain MainMainMsg, got:\n%s", output)
	}
}

func TestParseProto_FileNotFound(t *testing.T) {
	// Should log a warning and return empty protoFile
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	pf := parseProto("non-existent-file.proto")

	w.Close()
	os.Stderr = oldStderr
	io.Copy(io.Discard, r)

	if pf.pkg != "" {
		t.Errorf("expected empty protoFile, got package %q", pf.pkg)
	}
}

func TestExecute_FileNotFound(t *testing.T) {
	// Test that we try to resolve a missing file and log a warning correctly without crashing.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	execute([]string{"non-existent-file.proto"})

	w.Close()
	os.Stderr = oldStderr
	io.Copy(io.Discard, r)
}
