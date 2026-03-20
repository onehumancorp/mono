// GENERATED FILE – do not edit manually.
// Generated at build time from proto definitions in srcs/proto/.
// Run:  bazel build //srcs/proto:proto_types_ts

// ── ohc.common ──

/**
 * Summary: CommonRole is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: CommonAgentStatus is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: AgentAgent is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface AgentAgent {
  id?: string;
  role?: CommonRole;
  name?: string;
  status?: CommonAgentStatus;
  organizationId?: string;
}

/**
 * Summary: AgentAgentMessage is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: OrganizationRoleProfile is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrganizationRoleProfile {
  role?: CommonRole;
  basePrompt?: string;
  capabilities?: string[];
  contextInputs?: string[];
}

/**
 * Summary: OrganizationOrganization is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: OrganizationTeamMember is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: OrganizationOrganizationChart is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrganizationOrganizationChart {
  organization?: OrganizationOrganization;
  members?: OrganizationTeamMember[];
}

// ── ohc.billing ──

/**
 * Summary: BillingTokenUsage is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: BillingCostSummary is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface BillingCostSummary {
  organizationId?: string;
  totalCostUsd?: number;
  totalTokens?: number;
  projectedMonthlyUsd?: number;
  agents?: BillingAgentCostSummary[];
}

/**
 * Summary: BillingAgentCostSummary is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface BillingAgentCostSummary {
  agentId?: string;
  costUsd?: number;
  tokenUsed?: number;
}

/**
 * Summary: BillingBudgetAlert is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: OrchestrationAgent is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationAgent {
  id?: string;
  name?: string;
  role?: string;
  organizationId?: string;
  status?: string;
}

/**
 * Summary: OrchestrationMessage is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: OrchestrationMeetingRoom is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationMeetingRoom {
  id?: string;
  agenda?: string;
  participants?: string[];
  transcript?: OrchestrationMessage[];
}

/**
 * Summary: OrchestrationRegisterAgentRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationRegisterAgentRequest {
  agent?: OrchestrationAgent;
}

/**
 * Summary: OrchestrationRegisterAgentResponse is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationRegisterAgentResponse {
  success?: boolean;
}

/**
 * Summary: OrchestrationOpenMeetingRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationOpenMeetingRequest {
  meetingId?: string;
  agenda?: string;
  participants?: string[];
}

/**
 * Summary: OrchestrationPublishMessageRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationPublishMessageRequest {
  message?: OrchestrationMessage;
}

/**
 * Summary: OrchestrationPublishMessageResponse is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationPublishMessageResponse {
  success?: boolean;
}

/**
 * Summary: OrchestrationStreamMessagesRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationStreamMessagesRequest {
  agentId?: string;
}

/**
 * Summary: OrchestrationReasonRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationReasonRequest {
  prompt?: string;
}

/**
 * Summary: OrchestrationReasonResponse is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface OrchestrationReasonResponse {
  content?: string;
}

// ── ohc.api.v1 ──

/**
 * Summary: ApiMeetingRoom is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiMeetingRoom {
  id?: string;
  participants?: string[];
  transcript?: AgentAgentMessage[];
}

/**
 * Summary: ApiStatusCount is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiStatusCount {
  status?: string;
  count?: number;
}

/**
 * Summary: ApiDashboardSnapshot is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: ApiGetDashboardRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiGetDashboardRequest {
  organizationId?: string;
}

/**
 * Summary: ApiPostMessageRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiPostMessageRequest {
  message?: AgentAgentMessage;
}

/**
 * Summary: ApiPostMessageResponse is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiPostMessageResponse {
  snapshot?: ApiDashboardSnapshot;
}

/**
 * Summary: ApiSeedDashboardRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiSeedDashboardRequest {
  scenario?: string;
}

/**
 * Summary: ApiSeedDashboardResponse is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiSeedDashboardResponse {
  snapshot?: ApiDashboardSnapshot;
}

/**
 * Summary: ApiTrustAgreement is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: ApiB2BHandshakeRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiB2BHandshakeRequest {
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
}

/**
 * Summary: ApiIncident is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: ApiIncidentStatusRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiIncidentStatusRequest {
  incidentId?: string;
  status?: string;
  resolutionPlanId?: string;
}

/**
 * Summary: ApiComputeProfile is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiComputeProfile {
  roleId?: string;
  minVramGb?: number;
  preferredGpuType?: string;
  schedulingPriority?: number;
}

/**
 * Summary: ApiClusterStatus is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiClusterStatus {
  region?: string;
  status?: string;
  latencyMs?: number;
  availableNodes?: number;
}

/**
 * Summary: ApiPipeline is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: ApiPipelinePromoteRequest is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface ApiPipelinePromoteRequest {
  pipelineId?: string;
  approvedBy?: string;
}

// ── ohc.skills ──

/**
 * Summary: SkillsPhase is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface SkillsPhase {
  name?: string;
  description?: string;
  protocolDetails?: string;
}

/**
 * Summary: SkillsRoleBlueprint is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
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
 * Summary: SkillsSystemPromptBlueprint is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface SkillsSystemPromptBlueprint {
  coreDirective?: string;
  requiredContextVariables?: string[];
}

/**
 * Summary: SkillsTeamBlueprint is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface SkillsTeamBlueprint {
  id?: string;
  teamName?: string;
  systemPrompt?: SkillsSystemPromptBlueprint;
  roles?: SkillsRoleBlueprint[];
}

/**
 * Summary: SkillsSkillSet is undocumented.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: None
 */
export interface SkillsSkillSet {
  id?: string;
  name?: string;
  teamTemplates?: SkillsTeamBlueprint[];
}
