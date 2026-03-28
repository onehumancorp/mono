<div align="center">
  <h1>One Human Corp Examples</h1>
  <p><strong>Pre-configured, high-quality agent examples for the One Human Corp platform.</strong></p>
</div>

---

## Identity
The `examples` module provides a comprehensive suite of pre-configured, out-of-the-box reference implementations for AI agents, allowing developers to immediately test and observe the One Human Corp orchestration platform in action.

## Architecture
These examples are designed to practically demonstrate the platform's **Zero-Lock** paradigm. Production agents interact generically through abstraction layers, relying on `SPIFFE/SPIRE` for identity and Kubernetes Secrets for configuration injection. The specific `hello-world-agent` highlights how a generic provider interface is consumed by the application layer without hardcoded external API dependencies.

## Quick Start
Experience the platform in seconds with the "Hello World" agent. It leverages the `builtin` model for immediate feedback with **zero configuration** and **no external API keys**.
Run the compiled Go agent directly using our intuitive Bazel aliases:
```bash
bazelisk run //:hello-world
```
*Expected Output: A successful boot log and a friendly "Hello World" message.*

## Developer Workflow
The `examples` directory serves as a template and testing ground for new agent behaviors.
- **Build all examples:**
  ```bash
  bazelisk build //examples/...
  ```
- **Test all examples:**
  ```bash
  bazelisk test //examples/...
  ```

## Configuration
For local development, the `hello-world` uses the `builtin` model. For production deployment, you can deploy the raw Kubernetes Custom Resource Definition (CRD) to your local cluster:
```yaml
# examples/hello-world-agent/hello_world_agent.yaml
apiVersion: onehumancorp.com/v1alpha1
kind: Agent
metadata:
  name: hello-world
spec:
  role: "SOFTWARE_ENGINEER"
  model: "builtin"
  prompt: "You are a friendly Hello World agent..."
```
