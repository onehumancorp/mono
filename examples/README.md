<div align="center">
  <h1>One Human Corp Examples</h1>
  <p><strong>Pre-configured, high-quality agent examples for the One Human Corp platform.</strong></p>
</div>

---

## Identity
The `examples` module provides pre-configured, high-quality agent examples for the One Human Corp platform. It serves as a template and testing ground for new agent behaviors.

## Architecture
These examples demonstrate the **Zero-Lock** paradigm. Production agents rely on `SPIFFE/SPIRE` for identity and Kubernetes Secrets for configuration injection. The `hello-world-agent` specifically highlights how a generic provider interface is consumed by the application layer.

## Quick Start
Experience the platform in seconds. The `hello-world-agent` is designed to run perfectly out-of-the-box with **zero configuration** and **no external API keys** required. It leverages the `builtin` model for immediate feedback.

Run the compiled Go agent directly using our intuitive Bazel aliases:
```bash
bazelisk run //:hello-world
```
*Expected Output: A successful boot log and a friendly "Hello World" message.*

Alternatively, deploy the raw Kubernetes Custom Resource Definition (CRD) to your local cluster:
```yaml
# examples/hello_world_agent.yaml
apiVersion: onehumancorp.com/v1alpha1
kind: Agent
metadata:
  name: hello-world
spec:
  role: "SOFTWARE_ENGINEER"
  model: "builtin"
  prompt: "You are a friendly Hello World agent..."
```

## Developer Workflow
The examples directory serves as a template and testing ground for new agent behaviors.

- **Build all examples:**
  ```bash
  bazelisk build //examples/...
  ```
- **Test all examples:**
  ```bash
  bazelisk test //examples/...
  ```

## Configuration
No external API keys or complex configuration are required for the basic examples. Production variants may use Kubernetes Secrets for configuration injection.
