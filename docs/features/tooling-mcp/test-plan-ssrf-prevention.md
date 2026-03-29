# Test Plan: Integrations Registry SSRF Prevention


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** AI Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
This test plan details the testing strategy for the SSRF prevention feature in the Integrations Registry (`srcs/integrations/registry.go`). It ensures that malicious URLs cannot be used to establish connections to internal or restricted network resources.

## 2. Test Strategy
- **Unit Testing:** Implement comprehensive unit tests using Table-Driven tests in Go to validate the `validateURL()` function behavior across a wide range of URL types.

## 3. Test Cases
### 3.1 Unit Tests

| Test ID | Component | Description | Expected Result |
|---------|-----------|-------------|-----------------|
| UT-01 | `validateURL` | Valid External URL (e.g., `https://api.github.com`) | Success (No error) |
| UT-02 | `validateURL` | Loopback IP (`http://127.0.0.1`) | Error |
| UT-03 | `validateURL` | Loopback IPv6 (`http://[::1]`) | Error |
| UT-04 | `validateURL` | Localhost (`http://localhost`) | Error |
| UT-05 | `validateURL` | Private IP Class A (`http://10.0.0.1`) | Error |
| UT-06 | `validateURL` | Private IP Class B (`http://172.16.0.1`) | Error |
| UT-07 | `validateURL` | Private IP Class C (`http://192.168.1.1`) | Error |
| UT-08 | `validateURL` | Link-Local AWS IMDS (`http://169.254.169.254/latest/meta-data/`) | Error |
| UT-09 | `validateURL` | Unspecified IP (`http://0.0.0.0`) | Error |
| UT-10 | `validateURL` | Invalid URL Format (`htp://[::1]:80`) | Error |
| UT-11 | `Connect` | Call `Connect` with valid external URL | Success |
| UT-12 | `Connect` | Call `Connect` with Loopback IP | Error |
| UT-13 | `TestConnection`| Call `TestConnection` with Link-Local IP | Error |

## 4. Implementation Details
- The tests will be added to `srcs/integrations/registry_test.go`.
- The tests should achieve >95% coverage for the `validateURL` function and related branches in `Connect()` and `TestConnection()`.
- Tests will follow standard Go testing practices and inherit patterns from existing `*_test.go` files.

## 5. Edge Cases
- **DNS Rebinding:** Ensure the test suite considers DNS resolution and specifically tests `net.LookupIP` mocking.
- **Protocol Mismatch:** Testing malformed URL formats.
- **Missing host:** Validation of URL parsing logic when host is missing.
