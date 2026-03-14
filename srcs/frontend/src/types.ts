export type RoleProfile = {
  role: string;
  basePrompt: string;
  capabilities: string[];
  contextInputs: string[];
};

export type OrganizationMember = {
  id: string;
  name: string;
  role: string;
};

export type Organization = {
  id: string;
  name: string;
  domain: string;
  members: OrganizationMember[];
  roleProfiles: RoleProfile[];
};

export type MeetingMessage = {
  id: string;
  fromAgent: string;
  toAgent: string;
  type: string;
  content: string;
  meetingId: string;
  occurredAt: string;
};

export type MeetingRoom = {
  id: string;
  agenda: string;
  participants: string[];
  transcript: MeetingMessage[];
};

export type AgentCost = {
  agentID: string;
  model: string;
  tokenUsed: number;
  costUSD: number;
};

export type CostSummary = {
  organizationID: string;
  totalTokens: number;
  totalCostUSD: number;
  agents: AgentCost[];
};

export type StatusBucket = {
  status: string;
  count: number;
};

export type DashboardSnapshot = {
  organization: Organization;
  meetings: MeetingRoom[];
  costs: CostSummary;
  agents: {
    id: string;
    name: string;
    role: string;
    organizationId: string;
    status: string;
  }[];
  statuses: StatusBucket[];
  updatedAt: string;
};