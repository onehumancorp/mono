// GENERATED FILE – do not edit manually.
// Generated at build time from proto definitions in srcs/proto/.
// Run:  bazel build //srcs/proto:proto_types_ts

// ── ohc.common ──

/**
 * Summary: Provides CommonRole functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type CommonRole =
  | "CEO"
  | "DESIGNER"
  | "ENGINEERING_DIRECTOR"
  | "MARKETING_MANAGER"
  | "PRODUCT_MANAGER"
  | "QA_TESTER"
  | "ROLE_UNSPECIFIED"
  | "SECURITY_ENGINEER"
  | "SOFTWARE_ENGINEER"
;

/**
 * Summary: Provides CommonAgentStatus functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export type CommonAgentStatus =
  | "ACTIVE"
  | "BLOCKED"
  | "IDLE"
  | "IN_MEETING"
  | "STATUS_UNSPECIFIED"
;

// ── ohc.agent ──

/**
 * Summary: Provides AgentAgent functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface AgentAgent {
  id?: string;
  role?: CommonRole;
  name?: string;
  status?: CommonAgentStatus;
  organizationId?: string;
}

/**
 * Summary: Provides AgentAgentMessage functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface AgentAgentMessage {
  id?: string;
  fromAgentId?: string;
  toAgentId?: string;
  messageType?: string;
  content?: string;
  meetingId?: string;
  occurredAtUnix?: number;
}

// ── ohc.organization ──

/**
 * Summary: Provides OrganizationRoleProfile functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrganizationRoleProfile {
  role?: CommonRole;
  basePrompt?: string;
  capabilities?: string[];
  contextInputs?: string[];
}

/**
 * Summary: Provides OrganizationOrganization functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrganizationOrganization {
  id?: string;
  name?: string;
  domain?: string;
  ceoId?: string;
  createdAtUnix?: number;
  members?: OrganizationTeamMember[];
  roleProfiles?: OrganizationRoleProfile[];
}

/**
 * Summary: Provides OrganizationTeamMember functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrganizationTeamMember {
  id?: string;
  organizationId?: string;
  name?: string;
  role?: CommonRole;
  managerId?: string;
  isHuman?: boolean;
}

/**
 * Summary: Provides OrganizationOrganizationChart functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrganizationOrganizationChart {
  organization?: OrganizationOrganization;
  members?: OrganizationTeamMember[];
}

// ── ohc.billing ──

/**
 * Summary: Provides BillingTokenUsage functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface BillingTokenUsage {
  agentId?: string;
  organizationId?: string;
  model?: string;
  promptTokens?: number;
  completionTokens?: number;
  costUsd?: number;
  occurredAtUnix?: number;
}

/**
 * Summary: Provides BillingCostSummary functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface BillingCostSummary {
  organizationId?: string;
  totalCostUsd?: number;
  totalTokens?: number;
  projectedMonthlyUsd?: number;
  agents?: BillingAgentCostSummary[];
}

/**
 * Summary: Provides BillingAgentCostSummary functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface BillingAgentCostSummary {
  agentId?: string;
  costUsd?: number;
  tokenUsed?: number;
}

/**
 * Summary: Provides BillingBudgetAlert functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface BillingBudgetAlert {
  id?: string;
  organizationId?: string;
  thresholdUsd?: number;
  notifyAtPct?: number;
  triggered?: boolean;
}

// ── ohc.orchestration ──

/**
 * Summary: Provides OrchestrationAgent functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationAgent {
  id?: string;
  name?: string;
  role?: string;
  organizationId?: string;
  status?: string;
}

/**
 * Summary: Provides OrchestrationMessage functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationMessage {
  id?: string;
  fromAgent?: string;
  toAgent?: string;
  type?: string;
  content?: string;
  meetingId?: string;
  occurredAtUnix?: number;
}

/**
 * Summary: Provides OrchestrationMeetingRoom functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationMeetingRoom {
  id?: string;
  agenda?: string;
  participants?: string[];
  transcript?: OrchestrationMessage[];
}

/**
 * Summary: Provides OrchestrationRegisterAgentRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationRegisterAgentRequest {
  agent?: OrchestrationAgent;
}

/**
 * Summary: Provides OrchestrationRegisterAgentResponse functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationRegisterAgentResponse {
  success?: boolean;
}

/**
 * Summary: Provides OrchestrationOpenMeetingRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationOpenMeetingRequest {
  meetingId?: string;
  agenda?: string;
  participants?: string[];
}

/**
 * Summary: Provides OrchestrationPublishMessageRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationPublishMessageRequest {
  message?: OrchestrationMessage;
}

/**
 * Summary: Provides OrchestrationPublishMessageResponse functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationPublishMessageResponse {
  success?: boolean;
}

/**
 * Summary: Provides OrchestrationStreamMessagesRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationStreamMessagesRequest {
  agentId?: string;
}

/**
 * Summary: Provides OrchestrationReasonRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationReasonRequest {
  prompt?: string;
}

/**
 * Summary: Provides OrchestrationReasonResponse functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface OrchestrationReasonResponse {
  content?: string;
}

// ── ohc.api.v1 ──

/**
 * Summary: Provides ApiMeetingRoom functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiMeetingRoom {
  id?: string;
  participants?: string[];
  transcript?: AgentAgentMessage[];
}

/**
 * Summary: Provides ApiStatusCount functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiStatusCount {
  status?: string;
  count?: number;
}

/**
 * Summary: Provides ApiDashboardSnapshot functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiDashboardSnapshot {
  organization?: OrganizationOrganization;
  agents?: AgentAgent[];
  meetings?: ApiMeetingRoom[];
  costSummary?: BillingCostSummary;
  statuses?: ApiStatusCount[];
  updatedAt?: string;
}

/**
 * Summary: Provides ApiGetDashboardRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiGetDashboardRequest {
  organizationId?: string;
}

/**
 * Summary: Provides ApiPostMessageRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiPostMessageRequest {
  message?: AgentAgentMessage;
}

/**
 * Summary: Provides ApiPostMessageResponse functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiPostMessageResponse {
  snapshot?: ApiDashboardSnapshot;
}

/**
 * Summary: Provides ApiSeedDashboardRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiSeedDashboardRequest {
  scenario?: string;
}

/**
 * Summary: Provides ApiSeedDashboardResponse functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiSeedDashboardResponse {
  snapshot?: ApiDashboardSnapshot;
}

/**
 * Summary: Provides ApiTrustAgreement functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiTrustAgreement {
  id?: string;
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
  status?: string;
  createdAtUnix?: number;
}

/**
 * Summary: Provides ApiB2BHandshakeRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiB2BHandshakeRequest {
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
}

/**
 * Summary: Provides ApiIncident functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiIncident {
  id?: string;
  severity?: string;
  summary?: string;
  rootCauseAnalysis?: string;
  resolutionPlanId?: string;
  status?: string;
  createdAtUnix?: number;
}

/**
 * Summary: Provides ApiIncidentStatusRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiIncidentStatusRequest {
  incidentId?: string;
  status?: string;
  resolutionPlanId?: string;
}

/**
 * Summary: Provides ApiComputeProfile functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiComputeProfile {
  roleId?: string;
  minVramGb?: number;
  preferredGpuType?: string;
  schedulingPriority?: number;
}

/**
 * Summary: Provides ApiClusterStatus functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiClusterStatus {
  region?: string;
  status?: string;
  latencyMs?: number;
  availableNodes?: number;
}

/**
 * Summary: Provides ApiPipeline functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiPipeline {
  id?: string;
  name?: string;
  status?: string;
  branch?: string;
  stagingUrl?: string;
  initiatedBy?: string;
  createdAtUnix?: number;
  updatedAtUnix?: number;
}

/**
 * Summary: Provides ApiPipelinePromoteRequest functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface ApiPipelinePromoteRequest {
  pipelineId?: string;
  approvedBy?: string;
}

// ── ohc.skills ──

/**
 * Summary: Provides SkillsPhase functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface SkillsPhase {
  name?: string;
  description?: string;
  protocolDetails?: string;
}

/**
 * Summary: Provides SkillsRoleBlueprint functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface SkillsRoleBlueprint {
  id?: string;
  name?: string;
  roleArchetype?: CommonRole;
  level?: string;
  objective?: string;
  phases?: SkillsPhase[];
  constraints?: string[];
}

/**
 * Summary: Provides SkillsSystemPromptBlueprint functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface SkillsSystemPromptBlueprint {
  coreDirective?: string;
  requiredContextVariables?: string[];
}

/**
 * Summary: Provides SkillsTeamBlueprint functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface SkillsTeamBlueprint {
  id?: string;
  teamName?: string;
  systemPrompt?: SkillsSystemPromptBlueprint;
  roles?: SkillsRoleBlueprint[];
}

/**
 * Summary: Provides SkillsSkillSet functionality.
 * Intent: Supports the system's core functionality.
 * Params: See implementation
 * Returns: See implementation
 * Errors: Standard operational errors where applicable.
 * Side Effects: May interact with external systems or mutate internal state.
 */
export interface SkillsSkillSet {
  id?: string;
  name?: string;
  teamTemplates?: SkillsTeamBlueprint[];
}
