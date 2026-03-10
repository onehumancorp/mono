#!/usr/bin/env bash
set -euo pipefail

repo_name="${TEST_WORKSPACE:-mono}"
repo_root="${TEST_SRCDIR}/${repo_name}"

for file in common.proto agent.proto organization.proto billing.proto; do
  test -s "${repo_root}/srcs/proto/${file}"
done

grep -q 'service AgentOrchestration' "${repo_root}/srcs/proto/agent.proto"
grep -q 'service OrganizationService' "${repo_root}/srcs/proto/organization.proto"
grep -q 'service BillingService' "${repo_root}/srcs/proto/billing.proto"
