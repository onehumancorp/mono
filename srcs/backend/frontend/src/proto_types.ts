// GENERATED FILE – do not edit manually.
// Generated at build time from proto definitions in srcs/proto/.
// Run:  bazel build //srcs/proto:proto_types_ts

// ── ohc.common ──

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

export type CommonAgentStatus =
  | "ACTIVE"
  | "BLOCKED"
  | "IDLE"
  | "IN_MEETING"
  | "STATUS_UNSPECIFIED"
;

// ── ohc.agent ──

export interface AgentAgent {
  id?: string;
  role?: CommonRole;
  name?: string;
  status?: CommonAgentStatus;
  organizationId?: string;
}

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

export interface OrganizationRoleProfile {
  role?: CommonRole;
  basePrompt?: string;
  capabilities?: string[];
  contextInputs?: string[];
}

export interface OrganizationOrganization {
  id?: string;
  name?: string;
  domain?: string;
  ceoId?: string;
  createdAtUnix?: number;
  members?: OrganizationTeamMember[];
  roleProfiles?: OrganizationRoleProfile[];
}

export interface OrganizationTeamMember {
  id?: string;
  organizationId?: string;
  name?: string;
  role?: CommonRole;
  managerId?: string;
  isHuman?: boolean;
}

export interface OrganizationOrganizationChart {
  organization?: OrganizationOrganization;
  members?: OrganizationTeamMember[];
}

// ── ohc.billing ──

export interface BillingTokenUsage {
  agentId?: string;
  organizationId?: string;
  model?: string;
  promptTokens?: number;
  completionTokens?: number;
  costUsd?: number;
  occurredAtUnix?: number;
}

export interface BillingCostSummary {
  organizationId?: string;
  totalCostUsd?: number;
  totalTokens?: number;
  projectedMonthlyUsd?: number;
  agents?: BillingAgentCostSummary[];
}

export interface BillingAgentCostSummary {
  agentId?: string;
  costUsd?: number;
  tokenUsed?: number;
}

export interface BillingBudgetAlert {
  id?: string;
  organizationId?: string;
  thresholdUsd?: number;
  notifyAtPct?: number;
  triggered?: boolean;
}

// ── ohc.orchestration ──

export interface OrchestrationAgent {
  id?: string;
  name?: string;
  role?: string;
  organizationId?: string;
  status?: string;
  providerType?: string;
}

export interface OrchestrationMessage {
  id?: string;
  fromAgent?: string;
  toAgent?: string;
  type?: string;
  content?: string;
  meetingId?: string;
  occurredAtUnix?: number;
}

export interface OrchestrationMeetingRoom {
  id?: string;
  agenda?: string;
  participants?: string[];
  transcript?: OrchestrationMessage[];
}

export interface OrchestrationRegisterAgentRequest {
  agent?: OrchestrationAgent;
}

export interface OrchestrationRegisterAgentResponse {
  success?: boolean;
}

export interface OrchestrationOpenMeetingRequest {
  meetingId?: string;
  agenda?: string;
  participants?: string[];
}

export interface OrchestrationPublishMessageRequest {
  message?: OrchestrationMessage;
}

export interface OrchestrationPublishMessageResponse {
  success?: boolean;
}

export interface OrchestrationDelegateTaskRequest {
  fromAgentId?: string;
  toAgentId?: string;
  task?: OrchestrationMessage;
}

export interface OrchestrationDelegateTaskResponse {
  success?: boolean;
}

export interface OrchestrationSubTask {
  taskId?: string;
  targetRole?: string;
  instruction?: string;
  parentThreadId?: string;
}

export interface OrchestrationTokenEfficientContextSummarizationEvent {
  eventId?: string;
  agentId?: string;
  payload?: string;
}

export interface OrchestrationToolParameterAutoCorrectionEvent {
  eventId?: string;
  agentId?: string;
  payload?: string;
}

export interface OrchestrationStreamMessagesRequest {
  agentId?: string;
}

export interface OrchestrationReasonRequest {
  prompt?: string;
}

export interface OrchestrationReasonResponse {
  content?: string;
}

// ── ohc.api.v1 ──

export interface ApiMeetingRoom {
  id?: string;
  participants?: string[];
  transcript?: AgentAgentMessage[];
}

export interface ApiStatusCount {
  status?: string;
  count?: number;
}

export interface ApiDashboardSnapshot {
  organization?: OrganizationOrganization;
  agents?: AgentAgent[];
  meetings?: ApiMeetingRoom[];
  costSummary?: BillingCostSummary;
  statuses?: ApiStatusCount[];
  updatedAt?: string;
}

export interface ApiGetDashboardRequest {
  organizationId?: string;
}

export interface ApiPostMessageRequest {
  message?: AgentAgentMessage;
}

export interface ApiPostMessageResponse {
  snapshot?: ApiDashboardSnapshot;
}

export interface ApiSeedDashboardRequest {
  scenario?: string;
}

export interface ApiSeedDashboardResponse {
  snapshot?: ApiDashboardSnapshot;
}

export interface ApiTrustAgreement {
  id?: string;
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
  status?: string;
  createdAtUnix?: number;
}

export interface ApiB2BHandshakeRequest {
  partnerOrg?: string;
  partnerJwksUrl?: string;
  allowedRoles?: string[];
}

export interface ApiIncident {
  id?: string;
  severity?: string;
  summary?: string;
  rootCauseAnalysis?: string;
  resolutionPlanId?: string;
  status?: string;
  createdAtUnix?: number;
}

export interface ApiIncidentStatusRequest {
  incidentId?: string;
  status?: string;
  resolutionPlanId?: string;
}

export interface ApiComputeProfile {
  roleId?: string;
  minVramGb?: number;
  preferredGpuType?: string;
  schedulingPriority?: number;
}

export interface ApiClusterStatus {
  region?: string;
  status?: string;
  latencyMs?: number;
  availableNodes?: number;
}

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

export interface ApiPipelinePromoteRequest {
  pipelineId?: string;
  approvedBy?: string;
}

// ── ohc.skills ──

export interface SkillsPhase {
  name?: string;
  description?: string;
  protocolDetails?: string;
}

export interface SkillsRoleBlueprint {
  id?: string;
  name?: string;
  roleArchetype?: CommonRole;
  level?: string;
  objective?: string;
  phases?: SkillsPhase[];
  constraints?: string[];
}

export interface SkillsSystemPromptBlueprint {
  coreDirective?: string;
  requiredContextVariables?: string[];
}

export interface SkillsTeamBlueprint {
  id?: string;
  teamName?: string;
  systemPrompt?: SkillsSystemPromptBlueprint;
  roles?: SkillsRoleBlueprint[];
}

export interface SkillsSkillSet {
  id?: string;
  name?: string;
  teamTemplates?: SkillsTeamBlueprint[];
}

