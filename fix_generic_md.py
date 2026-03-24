import os
import re

def title_case(s):
    replacements = {
        "b2b": "B2B",
        "spiffe": "SPIFFE",
        "graphql": "GraphQL",
        "oidc": "OIDC",
        "api": "API",
        "ui": "UI",
        "dns": "DNS",
        "openapi": "OpenAPI",
        "mcp": "MCP"
    }
    words = s.split("-")
    resolved_words = []
    for word in words:
        if word.lower() in replacements:
            resolved_words.append(replacements[word.lower()])
        else:
            resolved_words.append(word.capitalize())
    return " ".join(resolved_words)

def title_case_struct(s):
    words = s.split("-")
    resolved_words = []
    for word in words:
        resolved_words.append(word.capitalize())
    return "".join(resolved_words)

def generate_specific_context(feature_id, doc_type):
    specifics = {
        "native-vision": {
            "cuj": "When an agent needs to interpret visual data from screenshots or video feeds, the system processes image buffers via the Multimodal Gateway without relying on external APIs.",
            "design": "Utilizes local visual encoders integrated with the MCP gateway to process byte streams and return structured spatial coordinates and bounding boxes. Integrates deeply with the Model Context Protocol (MCP) and Kubernetes operator to provide Native Vision capabilities seamlessly across all active Swarm Agents.",
            "test": "Inject a simulated screen buffer and assert the agent returns bounding boxes matching the deterministic visual payload."
        },
        "b2b-spiffe-federation": {
            "cuj": "When two independent OHC clusters need to collaborate, the Orchestrator initiates a SPIRE trust bundle exchange to issue federated SVIDs.",
            "design": "Extends the holding company CRD to specify federated trust domains, updating the SPIRE server configuration automatically via the K8s Operator. Integrates deeply with the Model Context Protocol (MCP) and Kubernetes operator to provide B2B SPIFFE Federation capabilities seamlessly across all active Swarm Agents.",
            "test": "Simulate a cross-cluster agent request and verify the Gateway properly validates the federated JWT-SVID."
        },
        "graphql-schema-introspection": {
            "cuj": "An agent autonomously queries the `/graphql` endpoint of an unknown service to dynamically generate required query payloads.",
            "design": "The MCP Gateway performs an introspection query on startup and maps the GraphQL schema directly to standard MCP tool schemas. Integrates deeply with the Model Context Protocol (MCP) and Kubernetes operator to provide GraphQL Schema Introspection capabilities seamlessly across all active Swarm Agents.",
            "test": "Point the gateway at a mock GraphQL server, verify the schema is correctly translated into tool specifications in the registry."
        },
        "fail-closed-dns-verification": {
            "cuj": "If an agent attempts to hit an external service and DNS fails or times out, the system defaults to blocking the request entirely.",
            "design": "Overrides the default Go network resolver inside agent pods to strictly enforce allowlists and return immediate NXDOMAIN for unregistered external IPs. Integrates deeply with the Model Context Protocol (MCP) and Kubernetes operator to provide Fail Closed DNS Verification capabilities seamlessly across all active Swarm Agents.",
            "test": "Simulate a DNS resolution failure and assert the pod receives a deterministic failure instead of hanging or falling back to a public resolver."
        },
        "oidc-issuer-verification": {
            "cuj": "When a human manager logs into the dashboard, the system securely validates their identity token against the configured OIDC provider.",
            "design": "Implements strict `aud` and `iss` claim checking in the Go backend middleware using the `go-oidc` package. Integrates deeply with the Model Context Protocol (MCP) and Kubernetes operator to provide OIDC Issuer Verification capabilities seamlessly across all active Swarm Agents.",
            "test": "Provide a mocked JWT with an invalid issuer and verify the backend returns a 401 Unauthorized."
        }
    }

    if feature_id in specifics:
        return specifics[feature_id][doc_type]

    if doc_type == "cuj":
        return f"When an AI agent or human operator needs to execute a task involving {title_case(feature_id)}, the system seamlessly provisions the necessary context, authenticates the request via SPIFFE, and processes the operation without breaking the established Zero-Lock toolchain or risking context bloat."
    elif doc_type == "design":
        return f"Integrates deeply with the Model Context Protocol (MCP) and Kubernetes operator to provide {title_case(feature_id)} capabilities seamlessly across all active Swarm Agents. It utilizes LangGraph Checkpointing backed by our native Kubernetes CSI Snapshotting."
    else:
        return f"Simulate an agent invoking the {title_case(feature_id)} functionality."


def write_cuj(filepath, feature_id, feature_name):
    content = f"""# CUJ: {feature_name}

**Persona:** Autonomous Agent / Human Manager
**Context:** Leveraging {feature_name} during standard operational workflows or cross-team collaboration.
**Success Metrics:** Task completion latency under 50ms, zero unauthorized access, and complete observability via the event log.

## 1. User Journey Overview
{generate_specific_context(feature_id, 'cuj')}

## 2. Detailed Step-by-Step Breakdown
| Step | Action | System Trigger | Resulting State | Verification |
|------|--------|----------------|-----------------|--------------|
| 1 | Action initiated by Agent/User | API call to Orchestration Hub | Request queued | Database Check |
| 2 | SPIFFE Authentication | Gateway verifies `AuthRole` | Request authorized | Log Check |
| 3 | Core Processing | The {feature_name} logic is executed | Operation completed | DB Check |
| 4 | Audit & Telemetry | Result appended to `events.jsonl` | Metric logged | DB Check |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Resource Exhaustion or Context Bloat
- **Detection**: The payload exceeds token limits or memory bounds.
- **Auto-Recovery**: The system immediately triggers context summarization or rate limiting, scaling back operations safely.
- **Manual Intervention**: The CEO can allocate more compute or force a termination.

### 3.2 Scenario: Authentication Failure
- **Detection**: Invalid or expired SVID presented during the operation.
- **Resolution**: Request is dropped instantly, and a security alert is forwarded to the CEO Dashboard.

## 4. UI/UX Details
- **Component IDs**: Rendered via the `FeatureViewer` and `OrgChartViewer`.
- **Visual Cues**: Agent status indicators show execution status.
- **Accessibility**: ARIA labels and keyboard navigation paths.

## 5. Security & Privacy
- Operations require explicit, short-lived SVID authentication.
- All actions are subject to strict Human-in-the-Loop gating for high-risk executions.
"""
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

def write_design(filepath, feature_id, feature_name, feature_struct_name):
    content = f"""# Design Document: {feature_name}

## 1. Executive Summary
**Objective:** Architect and implement {feature_name} to empower autonomous agents and human operators.
**Scope:** Integration within the core Orchestration Hub and the MCP Gateway, adhering to the Zero-Lock paradigm.

## 2. Architecture & Components
{generate_specific_context(feature_id, 'design')}

## 3. Data Flow
1. **Trigger:** The feature is invoked via Agent intent or a K8s event.
2. **Processing:** The Orchestration Hub routes the payload, verifying SPIFFE/SPIRE constraints.
3. **Execution:** The action is securely completed with all operations logged immutably.
4. **Result:** The system state is updated and the event is written to `events.jsonl`.

## 4. API & Data Models
```protobuf
message {feature_struct_name}Event {{
  string event_id = 1;
  string agent_id = 2;
  bytes payload = 3;
}}
```

## 5. Implementation Details
- Ensure strict JSON validation via `dec.DisallowUnknownFields()` when decoding related payloads.
- Maintain minimal memory overhead by avoiding O(N) string manipulations in hot paths.
- All K8s pods associated with this feature will enforce least privilege (e.g., `runAsNonRoot: true`, `readOnlyRootFilesystem: true`).
- Implement bounded memory growth by explicitly deleting map entries upon successful execution.
"""
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

def write_test(filepath, feature_id, feature_name):
    content = f"""# Test Plan: {feature_name}

## 1. Testing Strategy
Validate the end-to-end functionality, security boundaries, and performance constraints of the {feature_name} feature using hermetic, table-driven tests. Ensure we use the Database Seeder pattern to establish deterministic starting states.

## 2. Test Cases
### 2.1 E2E Integration Test: Standard Execution Flow
- **Setup:** A mock environment with a deterministic database state via `/api/dev/seed`.
- **Action:** {generate_specific_context(feature_id, 'test')}
- **Assertion:** Verify the operation completes successfully and the correct events are written to `events.jsonl`.

### 2.2 Edge Case: Strict Schema and Payload Validation
- **Setup:** Craft an invalid payload containing unknown JSON fields.
- **Action:** Submit the payload to the feature's API endpoint.
- **Assertion:** Verify the request is rejected immediately via `dec.DisallowUnknownFields()` and does not crash the server.

### 2.3 Edge Case: Memory and Resource Bounding
- **Setup:** Simulate a high-frequency barrage of requests.
- **Action:** Monitor the feature's map-based trackers and buffers.
- **Assertion:** Verify memory growth remains bounded and map entries are properly deleted after resolving tracked states.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST strictly exceed 95% for the corresponding Go packages.
- Tests will utilize lightweight dependency injection for fatal exit paths (`os.Exit`).
"""
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

if __name__ == "__main__":
    for root, dirs, files in os.walk('docs/features'):
        feature_id = os.path.basename(root)
        if feature_id == "features" or feature_id == "advanced-agentic-capabilities":
            continue

        feature_name = title_case(feature_id)
        feature_struct_name = title_case_struct(feature_id)

        cuj_path = os.path.join(root, 'cuj.md')
        design_path = os.path.join(root, 'design-doc.md')
        test_path = os.path.join(root, 'test-plan.md')

        # Only rewrite if it's generic, or write if it's missing.
        is_auto_generated = False

        # Try checking cuj
        if os.path.exists(cuj_path):
            with open(cuj_path, 'r', encoding='utf-8') as f:
                cuj_content = f.read()
                if "This feature enables the system to handle the complexity of" in cuj_content:
                    is_auto_generated = True
                if "When an AI agent or human operator needs to execute a task involving" in cuj_content:
                    is_auto_generated = True

        # Try checking design
        if os.path.exists(design_path):
            with open(design_path, 'r', encoding='utf-8') as f:
                design_content = f.read()
                if "Integrates deeply with the Model Context Protocol (MCP)" in design_content and "message " in design_content:
                    is_auto_generated = True

        # Try checking test
        if os.path.exists(test_path):
            with open(test_path, 'r', encoding='utf-8') as f:
                test_content = f.read()
                if "Validate the end-to-end functionality, security boundaries, and performance constraints of the" in test_content:
                    is_auto_generated = True

        if is_auto_generated or not os.path.exists(cuj_path) or not os.path.exists(design_path) or not os.path.exists(test_path):
            write_cuj(cuj_path, feature_id, feature_name)
            write_design(design_path, feature_id, feature_name, feature_struct_name)
            write_test(test_path, feature_id, feature_name)
