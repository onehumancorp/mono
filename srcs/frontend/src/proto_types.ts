// GENERATED FILE – do not edit manually.
// Generated at build time from proto definitions in srcs/proto/.
// Run:  bazel build //srcs/proto:proto_types_ts

// ── ohc.common ──

/**
 * @summary CommonRole encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary CommonAgentStatus encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary AgentAgent encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface AgentAgent {
  id?: string;
  role?: CommonRole;
  name?: string;
  status?: CommonAgentStatus;
  organizationId?: string;
}

/**
 * @summary AgentAgentMessage encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary OrganizationRoleProfile encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrganizationRoleProfile {
  role?: CommonRole;
  basePrompt?: string;
  capabilities?: string[];
  contextInputs?: string[];
}

/**
 * @summary OrganizationOrganization encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary OrganizationTeamMember encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary OrganizationOrganizationChart encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrganizationOrganizationChart {
  organization?: OrganizationOrganization;
  members?: OrganizationTeamMember[];
}

// ── ohc.billing ──

/**
 * @summary BillingTokenUsage encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary BillingCostSummary encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface BillingCostSummary {
  organizationId?: string;
  totalCostUsd?: number;
  totalTokens?: number;
  projectedMonthlyUsd?: number;
  agents?: BillingAgentCostSummary[];
}

/**
 * @summary BillingAgentCostSummary encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface BillingAgentCostSummary {
  agentId?: string;
  costUsd?: number;
  tokenUsed?: number;
}

/**
 * @summary BillingBudgetAlert encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary OrchestrationAgent encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationAgent {
  id?: string;
  name?: string;
  role?: string;
  organizationId?: string;
  status?: string;
}

/**
 * @summary OrchestrationMessage encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary OrchestrationMeetingRoom encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationMeetingRoom {
  id?: string;
  agenda?: string;
  participants?: string[];
  transcript?: OrchestrationMessage[];
}

/**
 * @summary OrchestrationRegisterAgentRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationRegisterAgentRequest {
  agent?: OrchestrationAgent;
}

/**
 * @summary OrchestrationRegisterAgentResponse encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationRegisterAgentResponse {
  success?: boolean;
}

/**
 * @summary OrchestrationOpenMeetingRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationOpenMeetingRequest {
  meetingId?: string;
  agenda?: string;
  participants?: string[];
}

/**
 * @summary OrchestrationPublishMessageRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationPublishMessageRequest {
  message?: OrchestrationMessage;
}

/**
 * @summary OrchestrationPublishMessageResponse encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationPublishMessageResponse {
  success?: boolean;
}

/**
 * @summary OrchestrationStreamMessagesRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationStreamMessagesRequest {
  agentId?: string;
}

/**
 * @summary OrchestrationReasonRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationReasonRequest {
  prompt?: string;
}

/**
 * @summary OrchestrationReasonResponse encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface OrchestrationReasonResponse {
  content?: string;
}

// ── ohc.api.v1 ──

/**
 * @summary ApiMeetingRoom encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiMeetingRoom {
  id?: string;
  participants?: string[];
  transcript?: AgentAgentMessage[];
}

/**
 * @summary ApiStatusCount encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiStatusCount {
  status?: string;
  count?: number;
}

/**
 * @summary ApiDashboardSnapshot encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary ApiGetDashboardRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiGetDashboardRequest {
  organizationId?: string;
}

/**
 * @summary ApiPostMessageRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiPostMessageRequest {
  message?: AgentAgentMessage;
}

/**
 * @summary ApiPostMessageResponse encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiPostMessageResponse {
  snapshot?: ApiDashboardSnapshot;
}

/**
 * @summary ApiSeedDashboardRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiSeedDashboardRequest {
  scenario?: string;
}

/**
 * @summary ApiSeedDashboardResponse encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiSeedDashboardResponse {
  snapshot?: ApiDashboardSnapshot;
}

/**
 * @summary ApiTrustAgreement encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary ApiB2BHandshakeRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiB2BHandshakeRequest {
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
}

/**
 * @summary ApiIncident encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary ApiIncidentStatusRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiIncidentStatusRequest {
  incidentId?: string;
  status?: string;
  resolutionPlanId?: string;
}

/**
 * @summary ApiComputeProfile encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiComputeProfile {
  roleId?: string;
  minVramGb?: number;
  preferredGpuType?: string;
  schedulingPriority?: number;
}

/**
 * @summary ApiClusterStatus encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiClusterStatus {
  region?: string;
  status?: string;
  latencyMs?: number;
  availableNodes?: number;
}

/**
 * @summary ApiPipeline encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary ApiPipelinePromoteRequest encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface ApiPipelinePromoteRequest {
  pipelineId?: string;
  approvedBy?: string;
}

// ── ohc.skills ──

/**
 * @summary SkillsPhase encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface SkillsPhase {
  name?: string;
  description?: string;
  protocolDetails?: string;
}

/**
 * @summary SkillsRoleBlueprint encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
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
 * @summary SkillsSystemPromptBlueprint encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface SkillsSystemPromptBlueprint {
  coreDirective?: string;
  requiredContextVariables?: string[];
}

/**
 * @summary SkillsTeamBlueprint encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface SkillsTeamBlueprint {
  id?: string;
  teamName?: string;
  systemPrompt?: SkillsSystemPromptBlueprint;
  roles?: SkillsRoleBlueprint[];
}

/**
 * @summary SkillsSkillSet encapsulates frontend UI state, type definitions, or functional API logic.
 * @param Object properties defining the shape of the interface.
 * @returns None
 * @throws None
 * @remarks Side Effects: None. Type definition only.
 */
export interface SkillsSkillSet {
  id?: string;
  name?: string;
  teamTemplates?: SkillsTeamBlueprint[];
}
