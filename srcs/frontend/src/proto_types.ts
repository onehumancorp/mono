// GENERATED FILE – do not edit manually.
// Generated at build time from proto definitions in srcs/proto/.
// Run:  bazel build //srcs/proto:proto_types_ts

// ── ohc.common ──
/**
 * Intent: Handles operations related to CommonRole.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to CommonAgentStatus.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to AgentAgent.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface AgentAgent {
  id?: string;
  role?: CommonRole;
  name?: string;
  status?: CommonAgentStatus;
  organizationId?: string;
}
/**
 * Intent: Handles operations related to AgentAgentMessage.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to OrganizationRoleProfile.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrganizationRoleProfile {
  role?: CommonRole;
  basePrompt?: string;
  capabilities?: string[];
  contextInputs?: string[];
}
/**
 * Intent: Handles operations related to OrganizationOrganization.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to OrganizationTeamMember.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to OrganizationOrganizationChart.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrganizationOrganizationChart {
  organization?: OrganizationOrganization;
  members?: OrganizationTeamMember[];
}

// ── ohc.billing ──
/**
 * Intent: Handles operations related to BillingTokenUsage.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to BillingCostSummary.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface BillingCostSummary {
  organizationId?: string;
  totalCostUsd?: number;
  totalTokens?: number;
  projectedMonthlyUsd?: number;
  agents?: BillingAgentCostSummary[];
}
/**
 * Intent: Handles operations related to BillingAgentCostSummary.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface BillingAgentCostSummary {
  agentId?: string;
  costUsd?: number;
  tokenUsed?: number;
}
/**
 * Intent: Handles operations related to BillingBudgetAlert.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to OrchestrationAgent.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationAgent {
  id?: string;
  name?: string;
  role?: string;
  organizationId?: string;
  status?: string;
}
/**
 * Intent: Handles operations related to OrchestrationMessage.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to OrchestrationMeetingRoom.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationMeetingRoom {
  id?: string;
  agenda?: string;
  participants?: string[];
  transcript?: OrchestrationMessage[];
}
/**
 * Intent: Handles operations related to OrchestrationRegisterAgentRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationRegisterAgentRequest {
  agent?: OrchestrationAgent;
}
/**
 * Intent: Handles operations related to OrchestrationRegisterAgentResponse.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationRegisterAgentResponse {
  success?: boolean;
}
/**
 * Intent: Handles operations related to OrchestrationOpenMeetingRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationOpenMeetingRequest {
  meetingId?: string;
  agenda?: string;
  participants?: string[];
}
/**
 * Intent: Handles operations related to OrchestrationPublishMessageRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationPublishMessageRequest {
  message?: OrchestrationMessage;
}
/**
 * Intent: Handles operations related to OrchestrationPublishMessageResponse.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationPublishMessageResponse {
  success?: boolean;
}
/**
 * Intent: Handles operations related to OrchestrationStreamMessagesRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationStreamMessagesRequest {
  agentId?: string;
}
/**
 * Intent: Handles operations related to OrchestrationReasonRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationReasonRequest {
  prompt?: string;
}
/**
 * Intent: Handles operations related to OrchestrationReasonResponse.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface OrchestrationReasonResponse {
  content?: string;
}

// ── ohc.api.v1 ──
/**
 * Intent: Handles operations related to ApiMeetingRoom.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiMeetingRoom {
  id?: string;
  participants?: string[];
  transcript?: AgentAgentMessage[];
}
/**
 * Intent: Handles operations related to ApiStatusCount.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiStatusCount {
  status?: string;
  count?: number;
}
/**
 * Intent: Handles operations related to ApiDashboardSnapshot.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to ApiGetDashboardRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiGetDashboardRequest {
  organizationId?: string;
}
/**
 * Intent: Handles operations related to ApiPostMessageRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiPostMessageRequest {
  message?: AgentAgentMessage;
}
/**
 * Intent: Handles operations related to ApiPostMessageResponse.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiPostMessageResponse {
  snapshot?: ApiDashboardSnapshot;
}
/**
 * Intent: Handles operations related to ApiSeedDashboardRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiSeedDashboardRequest {
  scenario?: string;
}
/**
 * Intent: Handles operations related to ApiSeedDashboardResponse.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiSeedDashboardResponse {
  snapshot?: ApiDashboardSnapshot;
}
/**
 * Intent: Handles operations related to ApiTrustAgreement.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to ApiB2BHandshakeRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiB2BHandshakeRequest {
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
}
/**
 * Intent: Handles operations related to ApiIncident.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to ApiIncidentStatusRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiIncidentStatusRequest {
  incidentId?: string;
  status?: string;
  resolutionPlanId?: string;
}
/**
 * Intent: Handles operations related to ApiComputeProfile.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiComputeProfile {
  roleId?: string;
  minVramGb?: number;
  preferredGpuType?: string;
  schedulingPriority?: number;
}
/**
 * Intent: Handles operations related to ApiClusterStatus.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiClusterStatus {
  region?: string;
  status?: string;
  latencyMs?: number;
  availableNodes?: number;
}
/**
 * Intent: Handles operations related to ApiPipeline.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to ApiPipelinePromoteRequest.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface ApiPipelinePromoteRequest {
  pipelineId?: string;
  approvedBy?: string;
}

// ── ohc.skills ──
/**
 * Intent: Handles operations related to SkillsPhase.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface SkillsPhase {
  name?: string;
  description?: string;
  protocolDetails?: string;
}
/**
 * Intent: Handles operations related to SkillsRoleBlueprint.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
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
 * Intent: Handles operations related to SkillsSystemPromptBlueprint.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface SkillsSystemPromptBlueprint {
  coreDirective?: string;
  requiredContextVariables?: string[];
}
/**
 * Intent: Handles operations related to SkillsTeamBlueprint.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface SkillsTeamBlueprint {
  id?: string;
  teamName?: string;
  systemPrompt?: SkillsSystemPromptBlueprint;
  roles?: SkillsRoleBlueprint[];
}
/**
 * Intent: Handles operations related to SkillsSkillSet.
 *
 * Params: None.
 *
 * Returns: Standard inferred return.
 *
 * Errors: Throws or returns errors if the operation fails.
 *
 * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.
 */
export interface SkillsSkillSet {
  id?: string;
  name?: string;
  teamTemplates?: SkillsTeamBlueprint[];
}
