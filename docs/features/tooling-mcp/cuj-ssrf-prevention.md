# CUJ: Integrations Registry SSRF Prevention


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** AI Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
This Customer User Journey (CUJ) describes the interactions and expected outcomes for preventing Server-Side Request Forgery (SSRF) vulnerabilities when an AI Agent or User interacts with the Integrations Registry.

## 2. Personas
- **Human CEO/Admin:** The user responsible for connecting new external integrations (e.g., GitHub, Slack) to the One Human Corp platform.
- **AI Agent:** An autonomous worker that may attempt to connect to external systems via the `Connect()` API or `TestConnection()` functionality.
- **Malicious User/Agent:** An entity attempting to leverage the Integrations Registry to probe internal network infrastructure or access restricted services.

## 3. User Journeys

### 3.1 Scenario A: Connecting to a Valid External Integration
1. **User Goal:** Connect the "GitHub" integration.
2. **Action:** The Admin provides the valid Base URL (e.g., `https://api.github.com`) and required credentials to the platform.
3. **System Response:** The system successfully resolves the domain name, confirms the IP addresses are public, validates the URL, and establishes the connection. The integration status updates to "connected".
4. **Outcome:** The GitHub integration is successfully connected and ready for use.

### 3.2 Scenario B: Attempting to Connect to a Loopback Address
1. **User Goal:** Attempt to access internal services running on the server.
2. **Action:** A malicious agent or user attempts to connect an integration using a loopback address (e.g., `http://127.0.0.1:8080/internal-api`).
3. **System Response:** The system parses the URL, identifies the loopback address, and immediately blocks the connection attempt.
4. **Outcome:** The system returns an error indicating that the URL resolves to a blocked IP address, and the connection is not established.

### 3.3 Scenario C: Attempting to Connect to a Private Network
1. **User Goal:** Probe the internal corporate network.
2. **Action:** An attacker attempts to connect an integration using a private IP address (e.g., `http://10.0.0.5:80/secret-service`).
3. **System Response:** The system parses the URL, identifies the private network class, and blocks the connection attempt.
4. **Outcome:** The connection is rejected, and the integration remains disconnected.

### 3.4 Scenario D: Attempting to Connect via AWS IMDS (Metadata Service)
1. **User Goal:** Extract cloud provider metadata (e.g., AWS EC2 Instance Metadata).
2. **Action:** An attacker uses the AWS IMDS IP address (`http://169.254.169.254/latest/meta-data/`).
3. **System Response:** The system detects the link-local IP address and blocks the request.
4. **Outcome:** The connection is rejected, protecting sensitive instance metadata.

### 3.5 Scenario E: Attempting to Connect to an Unresolvable Host (DNS Error)
1. **User Goal:** Bypass DNS-based validation by exploiting resolution failures.
2. **Action:** A user or agent provides a URL with a domain that cannot be resolved via DNS.
3. **System Response:** The system attempts to resolve the domain, encounters a DNS failure, and "fails closed."
4. **Outcome:** The connection is rejected with a "DNS resolution failed" error message.

## 4. Exceptions and Error Scenarios
- If the URL format is invalid, the system should return "invalid URL format".
- If the URL is missing a hostname, the system should return "URL must contain a host".

## 5. Implementation Details
- Uses Go's `net.ParseIP` and `net.LookupIP` to validate hostnames against an explicitly defined denylist of private, loopback, and reserved IP ranges (e.g., `127.0.0.0/8`, `10.0.0.0/8`, `169.254.0.0/16`).
- Fails closed on DNS resolution errors.

## 6. Edge Cases
- **DNS Rebinding Attacks:** The system should resolve the IP address and perform validation *before* establishing the connection to mitigate time-of-check to time-of-use (TOCTOU) race conditions.
- **IPv6 and Alternate Representations:** The system correctly handles IPv6 loopback (`::1`) and non-standard IP representations (e.g., octal or hex formats, if applicable) by relying on standard Go `net` package parsing.
- **Missing Protocol/Host:** The system rejects URLs missing a scheme or host entirely.
